package yamlreader

import (
	"fmt"
	"os"
	"urlshortner/internal/model"

	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) (*model.UrlStore, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	var config model.Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}
	store := &model.UrlStore{
		Redirects: make(map[string]string),
	}
	for _, item := range config.ShortURLs {
		store.Redirects[item.Path] = item.Url
	}
	return store, nil
}
