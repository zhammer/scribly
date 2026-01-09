package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ResendClient struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

func (r *ResendClient) SendEmail(ctx context.Context, email Email) error {
	resendEmail := map[string]interface{}{
		"from":    "Scribly <emails@scribly.ink>",
		"to":      []string{email.To.Email},
		"subject": email.Subject,
		"html":    email.Body,
	}

	body := bytes.Buffer{}
	if err := json.NewEncoder(&body).Encode(resendEmail); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.baseURL+"/emails", &body)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Bearer "+r.apiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("received non-200 response from Resend %d - '%s'", resp.StatusCode, string(respBody))
	}

	return nil
}

func NewResendClient(baseURL string, apiKey string) *ResendClient {
	return &ResendClient{
		client:  &http.Client{},
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}
