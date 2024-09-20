package download

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/VATSIM-UK/ukcp-srd-import/internal/airac"
)

type SrdDownloader struct {
	cycle              *airac.AiracCycle
	loadedCycleFile    *os.File
	latestDownloadFile *os.File
	downloadUrl        string
}

var (
	ErrFailedToScanLoadedCycle = errors.New("failed to scan loaded cycle file")
	ErrUpToDate                = errors.New("SRD is up to date")
	ErrLoadedChecksumFailed    = errors.New("failed to calculate checksum of loaded cycle file")
	ErrDownloadChecksumFailed  = errors.New("failed to calculate checksum of downloaded cycle file")
)

func NewSrdDownloader(cycle *airac.AiracCycle, fileDir, downloadUrl string) (*SrdDownloader, error) {
	loadedCycleFile, err := loadCycleFile(fileDir)
	if err != nil {
		return nil, err
	}
	latestDownloadFile, err := loadLatestDownloadFile(fileDir)
	if err != nil {
		return nil, err
	}

	return &SrdDownloader{
		cycle:              cycle,
		loadedCycleFile:    loadedCycleFile,
		latestDownloadFile: latestDownloadFile,
		downloadUrl:        downloadUrl,
	}, nil
}

func (d *SrdDownloader) Download(ctx context.Context, force bool) error {
	d.loadedCycleFile.Seek(0, 0)
	scanner := bufio.NewScanner(d.loadedCycleFile)
	scanner.Split(bufio.ScanWords)

	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return ErrFailedToScanLoadedCycle
	}

	loadedCycle := scanner.Text()
	fmt.Printf("Loaded cycle is %v\n", loadedCycle)
	fmt.Printf("Latest cycle is %v\n", d.cycle.Ident)

	// We already have the latest cycle
	if loadedCycle == d.cycle.Ident && !force {
		return ErrUpToDate
	}

	// So we need to download the latest cycle
	client := http.DefaultClient
	fmt.Printf("Downloading SRD file from %v\n", d.downloadUrl)
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
		return fmt.Errorf("Unable to download SRD, status code was %s", resp.Status)
	}

	// Write the response body into a temporary file
	tempFile, err := os.CreateTemp("/tmp", "ukcp-srd-import-download")
	if err != nil {
		return err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	fmt.Printf("Downloaded SRD file to %v\n", tempFile.Name())
	// Now we'll write the downloaded file to the latest download file
	tempFile.Seek(0, 0)
	d.latestDownloadFile.Truncate(0)
	d.latestDownloadFile.Seek(0, 0)
	_, err = io.Copy(d.latestDownloadFile, tempFile)

	if err != nil {
		return err
	}

	// And finally, write the airac ident to the loaded cycle file
	err = d.loadedCycleFile.Truncate(0)
	d.loadedCycleFile.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = d.loadedCycleFile.Write([]byte(d.cycle.Ident))
	if err != nil {
		return err
	}

	// Delete the temporary file
	err = os.Remove(tempFile.Name())
	if err != nil {
		return err
	}

	fmt.Printf("Finished SRD download for cycle %v\n", d.cycle.Ident)
	return d.completeDownload()
}

func (d *SrdDownloader) LatestFileLocation() string {
	return d.latestDownloadFile.Name()
}

func (d *SrdDownloader) completeDownload() error {
	err := d.loadedCycleFile.Close()
	if err != nil {
		return err
	}

	return d.latestDownloadFile.Close()
}

func filePath(dir, file string) string {
	return fmt.Sprintf("%s/%s", dir, file)
}

func loadCycleFile(dir string) (*os.File, error) {
	return os.OpenFile(filePath(dir, "ukcp-srd-import-loaded-cycle"), os.O_RDWR|os.O_CREATE, 0600)
}

func loadLatestDownloadFile(dir string) (*os.File, error) {
	return os.OpenFile(filePath(dir, "ukcp-srd-import-loaded-download.xlsx"), os.O_RDWR|os.O_CREATE, 0600)
}
