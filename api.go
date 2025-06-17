package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func getLiteLLMModels() ([]ModelInfo, error) {
	url := os.Getenv("LITELLM_URL") + "/models?return_wildcard_routes=false&include_model_access_groups=false"
	token := os.Getenv("LITELLM_APIKEY")
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("x-litellm-api-key", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response LiteLLMResp
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func callLiteLLM(model string, messages []Message) (*LiteLLMResponse, error) {
	fmt.Println("Calling LiteLLM with model:", model)
	// Prepare request
	reqBody := LiteLLMRequest{
		Model:    model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	// Make HTTP request
	req, _ := http.NewRequest("POST", litellmURL+"/v1/chat/completions", nil)
	req.Header.Set("x-litellm-api-key", os.Getenv("LITELLM_APIKEY"))
	req.Body = io.NopCloser(bytes.NewBuffer(jsonData))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LiteLLM API error: %s", string(body))
	}

	// Parse response
	var litellmResp LiteLLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&litellmResp); err != nil {
		return nil, err
	}

	return &litellmResp, nil
}
