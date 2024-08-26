package text

import (
	"io"

	"github.com/mmatongo/chew/internal/common"
	"gopkg.in/yaml.v3"
)

func ProcessYAML(r io.Reader, url string) ([]common.Chunk, error) {
	var data interface{}
	if err := yaml.NewDecoder(r).Decode(&data); err != nil {
		return nil, err
	}

	yamlStr, err := yaml.Marshal(data)
	if err != nil {
		return nil, err
	}

	return []common.Chunk{{Content: string(yamlStr), Source: url}}, nil
}
