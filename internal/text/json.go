package text

import (
	"encoding/json"
	"fmt"
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
		return nil, fmt.Errorf("failed to marshal json: %w", err)
	}

	return []common.Chunk{{Content: string(jsonStr), Source: url}}, nil
}
