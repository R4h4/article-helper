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
	"time"

	"github.com/cenkalti/backoff/v4"
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
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	// You might want to add more validation here, e.g., checking file type

	return file, nil
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
