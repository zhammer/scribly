package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SendgridClient struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

func (s *SendgridClient) SendEmail(ctx context.Context, email Email) error {
	sendgridEmail := map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{
				"to": []map[string]interface{}{
					{
						"email": email.To.Email,
						"name":  email.To.Username,
					},
				},
				"subject": email.Subject,
			},
		},
		"from": map[string]interface{}{
			"email": "emails@scribly.ink",
		},
		"content": []map[string]interface{}{
			{
				"type":  "text/html",
				"value": email.Body,
			},
		},
	}
	body := bytes.Buffer{}
	if err := json.NewEncoder(&body).Encode(sendgridEmail); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", s.baseURL+"/v3/mail/send", &body)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+s.apiKey)
	req.Header.Add("content-type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Received non-200 response from sendgrid %d - '%s'", resp.StatusCode, string(body))
	}

	return err
}

func NewSendgridClient(baseURL string, apiKey string) *SendgridClient {
	return &SendgridClient{
		client:  &http.Client{},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}
