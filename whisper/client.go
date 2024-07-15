package whisper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
)

const (
	maxFileSize = 25 * 1024 * 1024 // 25 MB, adjust as needed
)

// Config holds the configuration for the Whisper API client
type Config struct {
	APIEndpoint string
	APIKey      string
}

// Client is a Whisper API client
type Client struct {
	config Config
	client *http.Client
}

// NewClient creates a new Whisper API client
func NewClient(config Config) *Client {
	if config.APIEndpoint == "" {
		config.APIEndpoint = "https://api.openai.com/v1/audio/transcriptions"
	}

	return &Client{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TranscribeAudio transcribes an audio file using the Whisper API
func (c *Client) TranscribeAudio(ctx context.Context, filePath string) (string, error) {
	file, err := c.openAndValidateFile(filePath)
	if err != nil {
		return "", fmt.Errorf("file error: %w", err)
	}
	defer file.Close()

	body, contentType, err := c.createMultipartRequest(file)
	if err != nil {
		return "", fmt.Errorf("request creation error: %w", err)
	}

	var result string
	operation := func() error {
		resp, err := c.sendRequest(ctx, body, contentType)
		if err != nil {
			return fmt.Errorf("API request error: %w", err)
		}
		defer resp.Body.Close()

		result, err = c.parseResponse(resp)
		return err
	}

	err = backoff.Retry(operation, backoff.WithContext(backoff.NewExponentialBackOff(), ctx))
	if err != nil {
		return "", fmt.Errorf("transcription failed after retries: %w", err)
	}

	return result, nil
}

func (c *Client) openAndValidateFile(filePath string) (*os.File, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("error getting file info: %w", err)
	}

	// Check file size
	if fileInfo.Size() > maxFileSize {
		file.Close()
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxFileSize)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	allowedExtensions := []string{".mp3", ".mp4", ".mpeg", ".mpga", ".m4a", ".wav", ".webm"}
	if !contains(allowedExtensions, ext) {
		file.Close()
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	// Check file type using magic numbers
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("error reading file header: %w", err)
	}
	fileType := http.DetectContentType(buffer)

	allowedMimeTypes := []string{"audio/", "video/"}
	if !startsWithAny(fileType, allowedMimeTypes) {
		file.Close()
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}

	// Reset file pointer to the beginning
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("error resetting file pointer: %w", err)
	}

	return file, nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to check if a string starts with any of the given prefixes
func startsWithAny(s string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func (c *Client) createMultipartRequest(file *os.File) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		return nil, "", fmt.Errorf("error creating form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, "", fmt.Errorf("error copying file to form: %w", err)
	}

	err = writer.WriteField("model", "whisper-1")
	if err != nil {
		return nil, "", fmt.Errorf("error writing model field: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, "", fmt.Errorf("error closing multipart writer: %w", err)
	}

	return body, writer.FormDataContentType(), nil
}

func (c *Client) sendRequest(ctx context.Context, body *bytes.Buffer, contentType string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", c.config.APIEndpoint, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	req.Header.Set("Content-Type", contentType)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	return resp, nil
}

func (c *Client) parseResponse(resp *http.Response) (string, error) {
	var result struct {
		Text string `json:"text"`
	}

	err := json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	return result.Text, nil
}
