package download

import (
	"testing"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/airac"
)

func TestDownloadUrl(t *testing.T) {
	tests := []struct {
		name     string
		airac    *airac.AiracCycle
		expected string
	}{
		{
			name: "Valid AiracCycle",
			airac: &airac.AiracCycle{
				Ident: "2101",
			},
			expected: "https://www.nm.eurocontrol.int/RAD/additional doc/external_links/uk_srd/UK_Ireland_SRD_2101_notes.xlsx",
		},
		{
			name: "Empty AiracCycle Ident",
			airac: &airac.AiracCycle{
				Ident: "",
			},
			expected: "https://www.nm.eurocontrol.int/RAD/additional doc/external_links/uk_srd/UK_Ireland_SRD__notes.xlsx",
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
