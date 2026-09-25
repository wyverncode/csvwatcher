package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
