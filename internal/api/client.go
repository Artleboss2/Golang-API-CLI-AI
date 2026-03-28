package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	BaseURL                = "https://integrate.api.nvidia.com/v1"
	ChatCompletionEndpoint = "/chat/completions"
	RequestTimeout         = 120 * time.Second
	UserAgent              = "nvidia-nim-cli/1.0.0"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	Stream      bool      `json:"stream"`
	TopP        float64   `json:"top_p,omitempty"`
	Stop        []string  `json:"stop,omitempty"`
}

type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Delta        Message `json:"delta"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type StreamChunk struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

type APIError struct {
	StatusCode int
	Message    string
	Type       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Erreur API [%d] : %s", e.StatusCode, e.Message)
}

type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: RequestTimeout,
		},
		baseURL: BaseURL,
	}
}

func (c *Client) Complete(req *ChatRequest) (*ChatResponse, error) {
	req.Stream = false

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("erreur sérialisation JSON : %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+ChatCompletionEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("erreur création requête : %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("erreur réseau : %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erreur lecture réponse : %w", err)
	}

	if err := c.checkStatus(resp.StatusCode, respBody); err != nil {
		return nil, err
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("erreur parsing réponse : %w", err)
	}

	return &chatResp, nil
}

func (c *Client) StreamComplete(req *ChatRequest, tokenChan chan<- string, errChan chan<- error) {
	req.Stream = true

	body, err := json.Marshal(req)
	if err != nil {
		errChan <- fmt.Errorf("erreur sérialisation : %w", err)
		return
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+ChatCompletionEndpoint, bytes.NewBuffer(body))
	if err != nil {
		errChan <- fmt.Errorf("erreur création requête : %w", err)
		return
	}

	c.setHeaders(httpReq)
	httpReq.Header.Set("Accept", "text/event-stream")

	streamClient := &http.Client{}

	resp, err := streamClient.Do(httpReq)
	if err != nil {
		errChan <- fmt.Errorf("erreur réseau : %w", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		errChan <- c.checkStatus(resp.StatusCode, respBody)
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		if data == "[DONE]" {
			close(tokenChan)
			return
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 {
			token := chunk.Choices[0].Delta.Content
			if token != "" {
				tokenChan <- token
			}
		}
	}

	if err := scanner.Err(); err != nil {
		errChan <- fmt.Errorf("erreur lecture stream : %w", err)
		return
	}

	close(tokenChan)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent)
}

func (c *Client) checkStatus(statusCode int, body []byte) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}

	var apiErr struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}

	message := fmt.Sprintf("Code HTTP %d", statusCode)
	errType := "unknown"

	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error.Message != "" {
		message = apiErr.Error.Message
		errType = apiErr.Error.Type
	}

	switch statusCode {
	case 401:
		message = "Clé API invalide ou expirée. Utilisez 'nim auth' pour reconfigurer."
	case 403:
		message = "Accès refusé. Vérifiez vos permissions sur build.nvidia.com"
	case 404:
		message = "Modèle introuvable. Vérifiez l'ID du modèle."
	case 429:
		message = "Limite de taux dépassée. Attendez avant de réessayer."
	case 500, 502, 503:
		message = "Erreur serveur NVIDIA. Réessayez dans quelques instants."
	}

	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Type:       errType,
	}
}

func (c *Client) ValidateAPIKey(model string) error {
	req := &ChatRequest{
		Model: model,
		Messages: []Message{
			{Role: "user", Content: "Réponds uniquement par 'OK'."},
		},
		MaxTokens:   5,
		Temperature: 0.1,
		Stream:      false,
	}
	_, err := c.Complete(req)
	return err
}
