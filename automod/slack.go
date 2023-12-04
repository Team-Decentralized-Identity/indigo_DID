package automod

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type SlackWebhookBody struct {
	Text string `json:"text"`
}

// Sends a simple slack message to a channel via "incoming webhook".
//
// The slack incoming webhook must be already configured in the slack workplace.
func (e *Engine) SendSlackMsg(ctx context.Context, msg string) error {
	// loosely based on: https://golangcode.com/send-slack-messages-without-a-library/

	body, _ := json.Marshal(SlackWebhookBody{Text: msg})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.SlackWebhookURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if resp.StatusCode != 200 || buf.String() != "ok" {
		// TODO: in some cases print body? eg, if short and text
		return fmt.Errorf("failed slack webhook POST request. status=%d", resp.StatusCode)
	}
	return nil
}