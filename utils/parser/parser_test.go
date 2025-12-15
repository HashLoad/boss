package parser_test

import (
	"encoding/json"
	"testing"

	"github.com/hashload/boss/utils/parser"
)

func TestJSONMarshal_BasicStruct(t *testing.T) {
	type TestData struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	data := TestData{
		Name:    "test-package",
		Version: "1.0.0",
	}

	result, err := parser.JSONMarshal(data, false)
	if err != nil {
		t.Fatalf("JSONMarshal() error = %v", err)
	}

	if len(result) == 0 {
		t.Error("JSONMarshal() returned empty result")
	}

	// Verify it's valid JSON
	var parsed TestData
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Errorf("Result is not valid JSON: %v", err)
	}

	if parsed.Name != data.Name {
		t.Errorf("Name = %q, want %q", parsed.Name, data.Name)
	}
}

func TestJSONMarshal_SafeEncodingEnabled(t *testing.T) {
	type TestData struct {
		HTML string `json:"html"`
	}

	data := TestData{
		HTML: "<div>Test & Content</div>",
	}

	result, err := parser.JSONMarshal(data, true)
	if err != nil {
		t.Fatalf("JSONMarshal() error = %v", err)
	}

	resultStr := string(result)

	// With safeEncoding=true, <, >, & should NOT be escaped
	if contains(resultStr, "\\u003c") {
		t.Error("safeEncoding=true should not escape '<' as \\u003c")
	}
	if contains(resultStr, "\\u003e") {
		t.Error("safeEncoding=true should not escape '>' as \\u003e")
	}
	if contains(resultStr, "\\u0026") {
		t.Error("safeEncoding=true should not escape '&' as \\u0026")
	}

	// Should contain actual characters
	if !contains(resultStr, "<") {
		t.Error("safeEncoding=true should preserve '<' character")
	}
	if !contains(resultStr, ">") {
		t.Error("safeEncoding=true should preserve '>' character")
	}
	if !contains(resultStr, "&") {
		t.Error("safeEncoding=true should preserve '&' character")
	}
}

func TestJSONMarshal_SafeEncodingDisabled(t *testing.T) {
	type TestData struct {
		HTML string `json:"html"`
	}

	data := TestData{
		HTML: "<div>Test</div>",
	}

	result, err := parser.JSONMarshal(data, false)
	if err != nil {
		t.Fatalf("JSONMarshal() error = %v", err)
	}

	// With safeEncoding=false, characters may be escaped (standard Go behavior)
	// The result should still be valid JSON
	var parsed TestData
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Errorf("Result is not valid JSON: %v", err)
	}

	if parsed.HTML != data.HTML {
		t.Errorf("HTML = %q, want %q", parsed.HTML, data.HTML)
	}
}

func TestJSONMarshal_Indentation(t *testing.T) {
	type TestData struct {
		Name  string `json:"name"`
		Items []int  `json:"items"`
	}

	data := TestData{
		Name:  "test",
		Items: []int{1, 2, 3},
	}

	result, err := parser.JSONMarshal(data, false)
	if err != nil {
		t.Fatalf("JSONMarshal() error = %v", err)
	}

	resultStr := string(result)

	// Should contain tabs (indentation)
	if !contains(resultStr, "\t") {
		t.Error("JSONMarshal() should produce indented output with tabs")
	}

	// Should contain newlines
	if !contains(resultStr, "\n") {
		t.Error("JSONMarshal() should produce multi-line output")
	}
}

func TestJSONMarshal_EmptyStruct(t *testing.T) {
	type EmptyData struct{}

	data := EmptyData{}

	result, err := parser.JSONMarshal(data, true)
	if err != nil {
		t.Fatalf("JSONMarshal() error = %v", err)
	}

	if string(result) != "{}" {
		t.Errorf("JSONMarshal() = %q, want %q", string(result), "{}")
	}
}

func TestJSONMarshal_MapData(t *testing.T) {
	data := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	result, err := parser.JSONMarshal(data, true)
	if err != nil {
		t.Fatalf("JSONMarshal() error = %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Errorf("Result is not valid JSON: %v", err)
	}

	if parsed["key1"] != "value1" {
		t.Errorf("key1 = %q, want %q", parsed["key1"], "value1")
	}
}

func TestJSONMarshal_NilValue(t *testing.T) {
	result, err := parser.JSONMarshal(nil, true)
	if err != nil {
		t.Fatalf("JSONMarshal() error = %v", err)
	}

	if string(result) != "null" {
		t.Errorf("JSONMarshal(nil) = %q, want %q", string(result), "null")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && searchSubstring(s, substr)))
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
