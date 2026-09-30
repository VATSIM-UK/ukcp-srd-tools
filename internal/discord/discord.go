package discord

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

type DiscordNotificationData struct {
	WebhookURL string
	Content    string
}

func SendDiscordNotification(data DiscordNotificationData) error {
	payload := fmt.Sprintf(`{"content": "%s", "username": "UKCP SRD Tools"}`, data.Content)

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

func LoadWebhookURL() string {
	webhookUrl := os.Getenv("DISCORD_WEBHOOK_URL")

	if webhookUrl == "" {
		log.Error().Msg("DISCORD_WEBHOOK_URL environment variable is not set")
	}
	return webhookUrl
}
