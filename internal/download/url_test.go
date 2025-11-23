package download

import (
	"testing"

	"github.com/VATSIM-UK/ukcp-srd-tools/internal/airac"
)

func TestDownloadUrl(t *testing.T) {
	tests := []struct {
		name     string
		airac    *airac.AiracCycle
		expected string
	}{
		{
			name: "Valid AiracCycle 2101",
			airac: &airac.AiracCycle{
				Ident: "2101",
			},
			expected: "https://nats-uk.ead-it.com/cms-nats/export/sites/default/en/Publications/digital-datasets/SRD/AIRAC-01-2021.zip",
		},
		{
			name: "Valid AiracCycle 2512",
			airac: &airac.AiracCycle{
				Ident: "2512",
			},
			expected: "https://nats-uk.ead-it.com/cms-nats/export/sites/default/en/Publications/digital-datasets/SRD/AIRAC-12-2025.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DownloadUrl(tt.airac)
			if result != tt.expected {
				t.Errorf("DownloadUrl() = %v, want %v", result, tt.expected)
			}
		})
	}
}
