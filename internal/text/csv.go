package text

import (
	"encoding/csv"
	"io"
	"strings"

	"github.com/mmatongo/chew/internal/common"
)

func ProcessCSV(r io.Reader, url string) ([]common.Chunk, error) {
	csvReader := csv.NewReader(r)
	var records [][]string
	var err error

	records, err = csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	var chunks []common.Chunk
	for _, record := range records {
		chunks = append(chunks, common.Chunk{Content: strings.Join(record, ", "), Source: url})
	}

	return chunks, nil
}
