package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	stabilityCheckInterval = 100 * time.Millisecond
	stabilityTimeout       = 10 * time.Second
	watchInterval          = time.Second
)

func main() {
	inputDir := flag.String("input", "", "directory containing CSV files")
	outputDir := flag.String("output", "", "directory for converted JSON files")
	flag.Parse()

	if *inputDir == "" || *outputDir == "" {
		flag.Usage()
		os.Exit(2)
	}

	input, err := filepath.Abs(*inputDir)
	if err != nil {
		log.Fatalf("resolve input directory: %v", err)
	}
	output, err := filepath.Abs(*outputDir)
	if err != nil {
		log.Fatalf("resolve output directory: %v", err)
	}
	if err := os.MkdirAll(output, 0755); err != nil {
		log.Fatalf("create output directory: %v", err)
	}

	if err := processExisting(input, output); err != nil {
		log.Fatalf("process existing files: %v", err)
	}
	if err := watch(input, output); err != nil {
		log.Fatalf("watch input directory: %v", err)
	}
}

func processExisting(inputDir, outputDir string) error {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() && isCSV(entry.Name()) {
			processAndLog(filepath.Join(inputDir, entry.Name()), outputDir)
		}
	}
	return nil
}

func watch(inputDir, outputDir string) error {
	log.Printf("watching %s; output directory is %s", inputDir, outputDir)
	known := make(map[string]fileState)
	if err := rememberFiles(inputDir, known); err != nil {
		return err
	}
	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()
	for {
		<-ticker.C
		if err := scanForChanges(inputDir, outputDir, known); err != nil {
			log.Printf("watch scan failed: %v", err)
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

func scanForChanges(inputDir, outputDir string, known map[string]fileState) error {
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
		known[path] = current
		processAndLog(path, outputDir)
	}
	return nil
}

func processAndLog(sourcePath, outputDir string) {
	if err := waitForStableFile(sourcePath); err != nil {
		log.Printf("conversion failed for %s: %v", sourcePath, err)
		return
	}
	outputPath, err := convertFile(sourcePath, outputDir)
	if err != nil {
		log.Printf("conversion failed for %s: %v", sourcePath, err)
		return
	}
	log.Printf("converted %s -> %s", sourcePath, outputPath)
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

func waitForStableFile(path string) error {
	deadline := time.Now().Add(stabilityTimeout)
	var previous os.FileInfo
	for time.Now().Before(deadline) {
		current, err := os.Stat(path)
		if err != nil {
			return err
		}
		if previous != nil && current.Size() == previous.Size() && current.ModTime().Equal(previous.ModTime()) {
			return nil
		}
		previous = current
		time.Sleep(stabilityCheckInterval)
	}
	return fmt.Errorf("file did not become stable within %s", stabilityTimeout)
}

func isCSV(name string) bool {
	return strings.EqualFold(filepath.Ext(name), ".csv")
}
