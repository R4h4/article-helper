package editor

import (
	"encoding/json"
	"fmt"
)

const (
	openAIURL = "https://api.openai.com/v1/chat/completions"
)

type AIEditor struct {
	EditorAgent   Agent
	HeadlineAgent Agent
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

func NewAIEditor(apiKey string, modelUrl string) *AIEditor {
	if modelUrl == "" {
		modelUrl = openAIURL
	}

	return &AIEditor{
		EditorAgent:   NewOpenAIAgent(apiKey, "gpt-4o", modelUrl, EditorPrompt),
		HeadlineAgent: NewOpenAIAgent(apiKey, "gpt-3.5-turbo", modelUrl, HeadlinePrompt),
	}
}

func (e *AIEditor) EditAndSummarize(transcript string) (*EditorResponse, error) {
	result, err := e.EditorAgent.Process(transcript)
	if err != nil {
		return nil, fmt.Errorf("error processing with editor agent: %v", err)
	}

	var editorResp EditorResponse
	err = json.Unmarshal([]byte(result.(string)), &editorResp)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling editor response: %v", err)
	}

	return &editorResp, nil
}

func (e *AIEditor) CreateHeadline(summary string) (string, error) {
	result, err := e.HeadlineAgent.Process(summary)
	if err != nil {
		return "", fmt.Errorf("error processing with headline agent: %v", err)
	}

	return result.(string), nil
}
