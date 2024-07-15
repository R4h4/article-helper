package editor

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAIEditor(t *testing.T) {
	// Mock OpenAI API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Logf("Error reading request body: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		t.Logf("Received request body: %s", string(body))

		var req OpenAIRequest
		err = json.Unmarshal(body, &req)
		if err != nil {
			t.Logf("Error unmarshaling request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		var resp OpenAIResponse
		switch {
		case contains(req.Messages[1].Content, "You will be given a transcription of someone's speech."):
			resp = OpenAIResponse{
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{
					{
						Message: struct {
							Content string `json:"content"`
						}{
							Content: `{"cleaned_transcription": "This is a cleaned test transcript.", "summary": "This is a summary of the test transcript."}`,
						},
					},
				},
			}
		case contains(req.Messages[1].Content, "Based on the following summary, create a short, catchy headline"):
			resp = OpenAIResponse{
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{
					{
						Message: struct {
							Content string `json:"content"`
						}{
							Content: "Test_Transcript_Summary",
						},
					},
				},
			}
		default:
			t.Logf("Unrecognized request content: %s", req.Messages[1].Content)
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		respBody, err := json.Marshal(resp)
		if err != nil {
			t.Logf("Error marshaling response: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		t.Logf("Sending response: %s", string(respBody))
		w.Write(respBody)
	}))
	defer server.Close()

	// Use the mock server URL instead of the real OpenAI API URL
	testUrl := server.URL + "/v1/chat/completions"
	editor := NewAIEditor("test-api-key", testUrl)

	t.Run("EditAndSummarize", func(t *testing.T) {
		result, err := editor.EditAndSummarize("This is a test transcript.")
		if err != nil {
			t.Fatalf("EditAndSummarize failed: %v", err)
		}

		expectedCleanedTranscription := "This is a cleaned test transcript."
		if result.CleanedTranscription != expectedCleanedTranscription {
			t.Errorf("Expected cleaned transcription %q, got %q", expectedCleanedTranscription, result.CleanedTranscription)
		}

		expectedSummary := "This is a summary of the test transcript."
		if result.Summary != expectedSummary {
			t.Errorf("Expected summary %q, got %q", expectedSummary, result.Summary)
		}
	})

	t.Run("CreateHeadline", func(t *testing.T) {
		headline, err := editor.CreateHeadline("This is a summary of the test transcript.")
		if err != nil {
			t.Fatalf("CreateHeadline failed: %v", err)
		}

		expectedHeadline := "Test_Transcript_Summary"
		if headline != expectedHeadline {
			t.Errorf("Expected headline %q, got %q", expectedHeadline, headline)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
