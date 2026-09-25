package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReadCSVHandlesHeadersAndSpecialCharacters(t *testing.T) {
	input := "name,note\nAda,\"She said \"\"hello\"\" & waved\"\n"
	records, err := readCSV(strings.NewReader(input))
	if err != nil {
		t.Fatalf("readCSV() error = %v", err)
	}
	if got := records[0]["note"]; got != `She said "hello" & waved` {
		t.Fatalf("note = %q", got)
	}
}

func TestReadCSVRejectsMismatchedRows(t *testing.T) {
	_, err := readCSV(strings.NewReader("name,age\nAda\n"))
	if err == nil || !strings.Contains(err.Error(), "row 2 has 1 fields") {
		t.Fatalf("expected row length error, got %v", err)
	}
}

func TestConvertFileWritesFormattedJSON(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "people.csv")
	output := filepath.Join(dir, "out")
	if err := os.Mkdir(output, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("name,note\nAda,\"hello & goodbye\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	outputPath, err := convertFile(source, output)
	if err != nil {
		t.Fatalf("convertFile() error = %v", err)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "\"name\": \"Ada\"") || !strings.Contains(string(data), "\"note\": \"hello & goodbye\"") {
		t.Fatalf("unexpected JSON output:\n%s", data)
	}
	if !strings.HasPrefix(string(data), "[\n") {
		t.Fatalf("expected JSON array, got object:\n%s", data)
	}
}

func TestDirectoriesOverlap(t *testing.T) {
	root := t.TempDir()
	if !directoriesOverlap(root, root) {
		t.Fatal("expected identical directories to overlap")
	}
	if !directoriesOverlap(root, filepath.Join(root, "output")) {
		t.Fatal("expected nested directories to overlap")
	}
	if directoriesOverlap(root, filepath.Join(t.TempDir(), "output")) {
		t.Fatal("expected separate directories not to overlap")
	}
}

func TestScanRetriesFailedConversion(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "input")
	output := filepath.Join(root, "output")
	if err := os.MkdirAll(input, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(output, 0755); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(input, "retry.csv")
	if err := os.WriteFile(source, []byte("name,age\nAda\n"), 0644); err != nil {
		t.Fatal(err)
	}

	known := make(map[string]fileState)
	ctx := context.Background()
	if err := scanForChanges(ctx, input, output, known, time.Second); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(output, "retry.json")); !os.IsNotExist(err) {
		t.Fatal("failed conversion should not create output")
	}
	if err := os.WriteFile(source, []byte("name\nAda\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := scanForChanges(ctx, input, output, known, time.Second); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(output, "retry.json")); err != nil {
		t.Fatalf("expected retry output: %v", err)
	}
}
