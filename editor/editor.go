package editor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	openAIURL = "https://api.openai.com/v1/chat/completions"
)

type AIEditor struct {
	APIKey string
}

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type EditorResponse struct {
	CleanedTranscription string `json:"cleaned_transcription"`
	Summary              string `json:"summary"`
}

func NewAIEditor(apiKey string) *AIEditor {
	return &AIEditor{APIKey: apiKey}
}

func (e *AIEditor) EditAndSummarize(transcript string) (*EditorResponse, error) {
	query := fmt.Sprintf(EditorPrompt, transcript)

	requestBody := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "system", Content: "You are an AI assistant that edits and summarizes transcripts."},
			{Role: "user", Content: query},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %v", err)
	}

	req, err := http.NewRequest("POST", openAIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", e.APIKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request to OpenAI: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var openAIResp OpenAIResponse
	err = json.Unmarshal(body, &openAIResp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling OpenAI response: %v", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	var editorResp EditorResponse
	err = json.Unmarshal([]byte(openAIResp.Choices[0].Message.Content), &editorResp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling editor response: %v", err)
	}

	return &editorResp, nil
}
