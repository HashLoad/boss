package parser

import (
	"bytes"
	"encoding/json"
)

func JSONMarshal(v interface{}, safeEncoding bool) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "\t")

	if safeEncoding {
		b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
		b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}
	return b, err
}
