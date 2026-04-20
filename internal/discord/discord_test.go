package discord

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSendDiscordNotification_Success(t *testing.T) {
	require := require.New(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal("POST", r.Method)
		require.Equal("application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := SendDiscordNotification(DiscordNotificationData{
		WebhookURL: server.URL,
		Content:    "Test notification",
	})
	require.NoError(err)
}

func TestSendDiscordNotification_StatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantError  bool
	}{
		{"200 OK", 200, false},
		{"204 No Content", 204, false},
		{"299 Max Success", 299, false},
		{"300 Redirect", 300, true},
		{"400 Bad Request", 400, true},
		{"500 Server Error", 500, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			err := SendDiscordNotification(DiscordNotificationData{
				WebhookURL: server.URL,
				Content:    "Test",
			})

			if tt.wantError {
				require.Error(err)
			} else {
				require.NoError(err)
			}
		})
	}
}

func TestSendDiscordNotification_InvalidURL(t *testing.T) {
	require := require.New(t)

	err := SendDiscordNotification(DiscordNotificationData{
		WebhookURL: "invalid url",
		Content:    "Test",
	})
	require.Error(err)
}

func TestSendDiscordNotification_NetworkError(t *testing.T) {
	require := require.New(t)

	err := SendDiscordNotification(DiscordNotificationData{
		WebhookURL: "http://nonexistent.invalid",
		Content:    "Test",
	})
	require.Error(err)
}

func TestLoadWebhookURL_Success(t *testing.T) {
	require := require.New(t)

	webhookURL := "https://discord.com/api/webhooks/123/abc"
	t.Setenv("DISCORD_WEBHOOK_URL", webhookURL)

	result := LoadWebhookURL()
	require.Equal(webhookURL, result)
}

func TestLoadWebhookURL_Missing(t *testing.T) {
	require := require.New(t)

	t.Setenv("DISCORD_WEBHOOK_URL", "")

	result := LoadWebhookURL()
	require.Equal("", result)
}

func TestSendDiscordNotification_PayloadContent(t *testing.T) {
	require := require.New(t)

	var receivedPayload map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(err)

		err = json.Unmarshal(body, &receivedPayload)
		require.NoError(err)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	testContent := "Test import for cycle 2404"
	err := SendDiscordNotification(DiscordNotificationData{
		WebhookURL: server.URL,
		Content:    testContent,
	})
	require.NoError(err)

	require.Equal(testContent, receivedPayload["content"])
}
