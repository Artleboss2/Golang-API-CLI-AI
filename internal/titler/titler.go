package titler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	baseURL  = "https://integrate.api.nvidia.com/v1/chat/completions"
	maxChars = 6000
	timeout  = 20 * time.Second
)

var cleanRe = regexp.MustCompile(`(?i)^(titre\s*:?\s*|")+|("+)$`)

func Fallback() string {
	return fmt.Sprintf("Session_%s", time.Now().Format("20060102_150405"))
}

func Generate(apiKey, model, firstUserMsg string) (string, error) {
	if strings.TrimSpace(firstUserMsg) == "" {
		return Fallback(), nil
	}

	content := firstUserMsg
	if len(content) > maxChars {
		content = content[:maxChars]
	}

	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "Tu es un assistant qui génère des titres courts. Réponds UNIQUEMENT avec le titre, sans guillemets, sans ponctuation finale, sans préambule. Maximum 6 mots.",
			},
			{
				"role":    "user",
				"content": fmt.Sprintf("Génère un titre court (max 6 mots) pour une conversation qui commence par :\n\n%s", content),
			},
		},
		"max_tokens":  20,
		"temperature": 0.4,
		"stream":      false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return Fallback(), nil
	}

	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(body))
	if err != nil {
		return Fallback(), nil
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return Fallback(), nil
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != http.StatusOK {
		return Fallback(), nil
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(raw, &result); err != nil || len(result.Choices) == 0 {
		return Fallback(), nil
	}

	title := strings.TrimSpace(result.Choices[0].Message.Content)
	title = cleanRe.ReplaceAllString(title, "")
	title = strings.TrimSpace(title)

	if title == "" || len(strings.Fields(title)) > 10 {
		return Fallback(), nil
	}

	return title, nil
}
