package discord

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

type DiscordNotificationData struct {
	WebhookURL string
	Content string
}

func SendDiscordNotification(data DiscordNotificationData) error {
	payload := fmt.Sprintf(`{"content": "%s"}`, data.Content)
	
	req, err := http.NewRequest("POST", data.WebhookURL, strings.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to send Discord notification, status code: %d", resp.StatusCode)
	}

	log.Info().Msg("Discord notification sent successfully")
	return nil
}