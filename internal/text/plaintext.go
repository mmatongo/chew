package text

import (
	"io"

	"github.com/mmatongo/chew/v1/internal/common"
)

func ProcessText(r io.Reader, url string) ([]common.Chunk, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if len(content) == 0 {
		return nil, nil
	}

	return []common.Chunk{{Content: string(content), Source: url}}, nil
}
