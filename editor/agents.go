package editor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Agent interface {
	Process(input string) (interface{}, error)
}

type OpenAIAgent struct {
	APIKey string
	URL    string
	Model  string
	Prompt string
}

func NewOpenAIAgent(apiKey, model, url, prompt string) *OpenAIAgent {
	return &OpenAIAgent{
		APIKey: apiKey,
		URL:    url,
		Model:  model,
		Prompt: prompt,
	}
}

func (a *OpenAIAgent) Process(input string) (interface{}, error) {
	query := fmt.Sprintf(a.Prompt, input)

	requestBody := OpenAIRequest{
		Model: a.Model,
		Messages: []Message{
			{Role: "system", Content: "You are editorAI. A large language model tasks with transcribing, correcting and summarizing text content."},
			{Role: "user", Content: query},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %v", err)
	}

	req, err := http.NewRequest("POST", a.URL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.APIKey))

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
		return nil, fmt.Errorf("error unmarshaling OpenAI response: %v. Response body: %s", err, string(body))
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response. Response body: %s", string(body))
	}

	return openAIResp.Choices[0].Message.Content, nil
}
