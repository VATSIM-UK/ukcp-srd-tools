package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/airac"
)

type SrdDownloader struct {
	cycle              *airac.AiracCycle
	loadedCycle        loadedAirac
	latestDownloadFile *os.File
	downloadUrl        string
}

type loadedAirac interface {
	Ident() string
	Is(ident string) bool
}

var (
	ErrFailedToScanLoadedCycle = errors.New("failed to scan loaded cycle file")
	ErrUpToDate                = errors.New("SRD is up to date")
	ErrLoadedChecksumFailed    = errors.New("failed to calculate checksum of loaded cycle file")
	ErrDownloadChecksumFailed  = errors.New("failed to calculate checksum of downloaded cycle file")
)

func NewSrdDownloader(cycle *airac.AiracCycle, loadedCycle loadedAirac, fileDir, downloadUrl string) (*SrdDownloader, error) {
	latestDownloadFile, err := loadLatestDownloadFile(fileDir)
	if err != nil {
		return nil, err
	}

	return &SrdDownloader{
		cycle:              cycle,
		loadedCycle:        loadedCycle,
		latestDownloadFile: latestDownloadFile,
		downloadUrl:        downloadUrl,
	}, nil
}

func (d *SrdDownloader) Download(ctx context.Context, force bool) error {
	log.Debug().Msg("Starting SRD download")
	log.Debug().Msg("Checking if SRD is up to date")
	log.Debug().Msgf("Loaded cycle is %v", d.loadedCycle.Ident())
	log.Debug().Msgf("Latest cycle is %v", d.cycle.Ident)

	// We already have the latest cycle
	if d.loadedCycle.Is(d.cycle.Ident) && !force {
		log.Info().Msg("SRD is up to date")
		return ErrUpToDate
	}

	// So we need to download the latest cycle
	client := http.DefaultClient
	log.Debug().Msgf("Downloading SRD file from %v", d.downloadUrl)
	req, err := http.NewRequestWithContext(ctx, "GET", d.downloadUrl, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("unable to download SRD, status code was %s", resp.Status)
		log.Error().Msg(msg)
		return errors.New(msg)
	}

	// Write the response body into a temporary file
	tempFile, err := os.CreateTemp("/tmp", "ukcp-srd-import-download")
	if err != nil {
		return err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	log.Debug().Msgf("Downloaded SRD file to %v", tempFile.Name())
	// Now we'll write the downloaded file to the latest download file
	tempFile.Seek(0, 0)
	d.latestDownloadFile.Truncate(0)
	d.latestDownloadFile.Seek(0, 0)
	_, err = io.Copy(d.latestDownloadFile, tempFile)

	if err != nil {
		return err
	}

	// Delete the temporary file
	err = os.Remove(tempFile.Name())
	if err != nil {
		return err
	}

	log.Info().Msg("finished SRD download")
	return d.completeDownload()
}

func (d *SrdDownloader) LatestFileLocation() string {
	return d.latestDownloadFile.Name()
}

func (d *SrdDownloader) completeDownload() error {
	return d.latestDownloadFile.Close()
}

func filePath(dir, file string) string {
	return fmt.Sprintf("%s/%s", dir, file)
}

func loadLatestDownloadFile(dir string) (*os.File, error) {
	return os.OpenFile(filePath(dir, "ukcp-srd-import-loaded-download.xlsx"), os.O_RDWR|os.O_CREATE, 0600)
}
