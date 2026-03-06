package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func Send(url string, payload interface{}) error {
	data, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("Failed to marshal payload: %w", err)
	}

	body := bytes.NewReader(data)

	resp, err := http.Post(url, "application/json", body)

	if err != nil {
		return fmt.Errorf("http post failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}
	return nil
}
