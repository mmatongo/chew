package text

import (
	"encoding/json"
	"io"

	"github.com/mmatongo/chew/internal/common"
)

func ProcessJSON(r io.Reader, url string) ([]common.Chunk, error) {
	var data interface{}
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return nil, err
	}

	jsonStr, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}

	return []common.Chunk{{Content: string(jsonStr), Source: url}}, nil
}
