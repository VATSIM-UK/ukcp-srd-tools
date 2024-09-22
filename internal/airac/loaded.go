package airac

import (
	"bufio"
	"fmt"
	"os"
)

type LoadedAirac struct {
	ident string

	loadedCycleFile *os.File
}

func NewLoadedAirac(dir string) (*LoadedAirac, error) {
	loadedCycleFile, err := loadCycleFile(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to load loaded cycle file: %w", err)
	}

	// Scan the file for the loaded cycle
	scanner := bufio.NewScanner(loadedCycleFile)
	scanner.Split(bufio.ScanWords)

	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan loaded cycle file: %w", err)
	}

	return &LoadedAirac{
		ident:           scanner.Text(),
		loadedCycleFile: loadedCycleFile,
	}, nil
}

func (l *LoadedAirac) Ident() string {
	return l.ident
}

func (l *LoadedAirac) Is(ident string) bool {
	return l.ident == ident
}

func (l *LoadedAirac) Close() error {
	return l.loadedCycleFile.Close()
}

func (l *LoadedAirac) Set(cycle *AiracCycle) error {
	l.loadedCycleFile.Seek(0, 0)
	l.loadedCycleFile.Truncate(0)
	_, err := l.loadedCycleFile.WriteString(cycle.Ident)
	return err
}

func filePath(dir, file string) string {
	return fmt.Sprintf("%s/%s", dir, file)
}

func loadCycleFile(dir string) (*os.File, error) {
	return os.OpenFile(filePath(dir, "ukcp-srd-import-loaded-cycle"), os.O_RDWR|os.O_CREATE, 0600)
}
