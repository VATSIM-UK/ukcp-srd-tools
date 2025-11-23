package download

import (
	"fmt"

	"github.com/VATSIM-UK/ukcp-srd-tools/internal/airac"
)

func DownloadUrl(airac *airac.AiracCycle) string {
	return fmt.Sprintf("https://nats-uk.ead-it.com/cms-nats/export/sites/default/en/Publications/digital-datasets/SRD/AIRAC-%s-20%s.zip", airac.MonthString(), airac.YearString())
}
