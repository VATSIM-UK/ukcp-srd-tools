package download

import (
	"fmt"

	"github.com/VATSIM-UK/ukcp-srd-tools/internal/airac"
)

func DownloadUrl(airac *airac.AiracCycle) string {
	return fmt.Sprintf("https://www.nm.eurocontrol.int/RAD/additional doc/external_links/uk_srd/UK_Ireland_SRD_%s_notes.xlsx", airac.Ident)
}
