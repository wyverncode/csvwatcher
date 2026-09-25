package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	defaultStabilityCheckInterval = 100 * time.Millisecond
	defaultStabilityTimeout       = 10 * time.Second
	defaultWatchInterval          = time.Second
)

func main() {
	inputDir := flag.String("input", "", "directory containing CSV files (required)")
	outputDir := flag.String("output", "", "directory for converted JSON files (required)")
	gui := flag.Bool("gui", false, "launch the cross-platform desktop interface")
	watchInterval := flag.Duration("interval", defaultWatchInterval, "how often to scan for changed files")
	stabilityTimeout := flag.Duration("stability-timeout", defaultStabilityTimeout, "maximum time to wait for a file to stop changing")
	once := flag.Bool("once", false, "convert existing files and exit without watching")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s --input DIR --output DIR [options]\n\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "Convert CSV files to formatted JSON and watch for changes.")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *gui {
		runGUI()
		return
	}
	if *inputDir == "" || *outputDir == "" {
		flag.Usage()
		os.Exit(2)
	}
	if *watchInterval <= 0 || *stabilityTimeout <= 0 {
		log.Fatal("intervals must be greater than zero")
	}

	input, err := filepath.Abs(*inputDir)
	if err != nil {
		log.Fatalf("resolve input directory: %v", err)
	}
	output, err := filepath.Abs(*outputDir)
	if err != nil {
		log.Fatalf("resolve output directory: %v", err)
	}
	if directoriesOverlap(input, output) {
		log.Fatalf("input and output directories must not overlap: %s and %s", input, output)
	}
	info, err := os.Stat(input)
	if err != nil {
		log.Fatalf("inspect input directory: %v", err)
	}
	if !info.IsDir() {
		log.Fatalf("input path is not a directory: %s", input)
	}
	if err := os.MkdirAll(output, 0755); err != nil {
		log.Fatalf("create output directory: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := processExisting(ctx, input, output, *stabilityTimeout); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		log.Fatalf("process existing files: %v", err)
	}
	if *once {
		return
	}

	if err := watch(ctx, input, output, *watchInterval, *stabilityTimeout); err != nil {
		log.Fatalf("watch input directory: %v", err)
	}
}

func processExisting(ctx context.Context, inputDir, outputDir string, stabilityTimeout time.Duration) error {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() && isCSV(entry.Name()) {
			if !processAndLog(ctx, filepath.Join(inputDir, entry.Name()), outputDir, stabilityTimeout) && ctx.Err() != nil {
				return ctx.Err()
			}
		}
	}
	return nil
}

func watch(ctx context.Context, inputDir, outputDir string, watchInterval, stabilityTimeout time.Duration) error {
	log.Printf("watching %s; output directory is %s", inputDir, outputDir)
	known := make(map[string]fileState)
	if err := rememberFiles(inputDir, known); err != nil {
		return err
	}
	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Printf("stopped watching")
			return nil
		case <-ticker.C:
			if err := scanForChanges(ctx, inputDir, outputDir, known, stabilityTimeout); err != nil {
				log.Printf("watch scan failed: %v", err)
			}
		}
	}
}

type fileState struct {
	size    int64
	modTime time.Time
}

func rememberFiles(inputDir string, known map[string]fileState) error {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !isCSV(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		known[filepath.Join(inputDir, entry.Name())] = fileState{size: info.Size(), modTime: info.ModTime()}
	}
	return nil
}

func scanForChanges(ctx context.Context, inputDir, outputDir string, known map[string]fileState, stabilityTimeout time.Duration) error {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !isCSV(entry.Name()) {
			continue
		}
		path := filepath.Join(inputDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			log.Printf("inspect %s failed: %v", path, err)
			continue
		}
		current := fileState{size: info.Size(), modTime: info.ModTime()}
		if previous, exists := known[path]; exists && previous == current {
			continue
		}
		if processAndLog(ctx, path, outputDir, stabilityTimeout) {
			known[path] = current
		}
	}
	return nil
}

func processAndLog(ctx context.Context, sourcePath, outputDir string, stabilityTimeout time.Duration) bool {
	if err := waitForStableFile(ctx, sourcePath, stabilityTimeout); err != nil {
		log.Printf("conversion failed for %s: %v", sourcePath, err)
		return false
	}
	outputPath, err := convertFile(sourcePath, outputDir)
	if err != nil {
		log.Printf("conversion failed for %s: %v", sourcePath, err)
		return false
	}
	log.Printf("converted %s -> %s", sourcePath, outputPath)
	return true
}

func convertFile(sourcePath, outputDir string) (string, error) {
	file, err := os.Open(sourcePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	records, err := readCSV(file)
	if err != nil {
		return "", err
	}
	outputPath := filepath.Join(outputDir, strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))+".json")
	data, err := formatJSON(records)
	if err != nil {
		return "", fmt.Errorf("marshal JSON: %w", err)
	}
	if err := writeAtomic(outputPath, data); err != nil {
		return "", err
	}
	return outputPath, nil
}

func formatJSON(records []map[string]string) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(records); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func readCSV(reader io.Reader) ([]map[string]string, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	headers, err := csvReader.Read()
	if err == io.EOF {
		return []map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read headers: %w", err)
	}
	if len(headers) == 0 {
		return nil, errors.New("CSV has no headers")
	}
	seen := make(map[string]struct{}, len(headers))
	for i, header := range headers {
		if header == "" {
			return nil, fmt.Errorf("header %d is empty", i+1)
		}
		if _, exists := seen[header]; exists {
			return nil, fmt.Errorf("duplicate header %q", header)
		}
		seen[header] = struct{}{}
	}

	records := make([]map[string]string, 0)
	for row := 2; ; row++ {
		values, err := csvReader.Read()
		if err == io.EOF {
			return records, nil
		}
		if err != nil {
			return nil, fmt.Errorf("read row %d: %w", row, err)
		}
		if len(values) != len(headers) {
			return nil, fmt.Errorf("row %d has %d fields; expected %d", row, len(values), len(headers))
		}
		record := make(map[string]string, len(headers))
		for i, header := range headers {
			record[header] = values[i]
		}
		records = append(records, record)
	}
}

func writeAtomic(path string, data []byte) error {
	temp, err := os.CreateTemp(filepath.Dir(path), ".csvwatcher-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary output: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)

	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return fmt.Errorf("write temporary output: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temporary output: %w", err)
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("replace output: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("publish output: %w", err)
	}
	return nil
}

func waitForStableFile(ctx context.Context, path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var previous os.FileInfo
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return err
		}
		current, err := os.Stat(path)
		if err != nil {
			return err
		}
		if previous != nil && current.Size() == previous.Size() && current.ModTime().Equal(previous.ModTime()) {
			return nil
		}
		previous = current
		timer := time.NewTimer(defaultStabilityCheckInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return fmt.Errorf("file did not become stable within %s", timeout)
}

func isCSV(name string) bool {
	return strings.EqualFold(filepath.Ext(name), ".csv")
}

func directoriesOverlap(first, second string) bool {
	first = filepath.Clean(first)
	second = filepath.Clean(second)
	return pathWithin(first, second) || pathWithin(second, first)
}

func pathWithin(parent, candidate string) bool {
	relative, err := filepath.Rel(parent, candidate)
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
}
