package parser

import (
	"bytes"
	"encoding/json"
)

// JSONMarshal marshals a value to JSON with optional safe encoding
func JSONMarshal(v any, safeEncoding bool) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "\t")

	if safeEncoding {
		b = bytes.ReplaceAll(b, []byte("\\u003c"), []byte("<"))
		b = bytes.ReplaceAll(b, []byte("\\u003e"), []byte(">"))
		b = bytes.ReplaceAll(b, []byte("\\u0026"), []byte("&"))
	}
	return b, err
}
