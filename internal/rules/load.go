package rules

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed catalog.yaml
var catalogBytes []byte

func Load() (*Catalog, error) {
	var c Catalog
	if err := yaml.Unmarshal(catalogBytes, &c); err != nil {
		return nil, fmt.Errorf("parse catalog: %w", err)
	}
	return &c, nil
}
