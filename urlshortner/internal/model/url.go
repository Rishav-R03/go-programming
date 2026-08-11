package model

type URLMapping struct {
	Path string `yaml:"path"`
	Url  string `yaml:"url"`
}

type Config struct {
	ShortURLs []URLMapping `yaml:"short_urls"`
}

type UrlStore struct {
	Redirects map[string]string
}

func (s *UrlStore) GetURL(path string) (string, bool) {
	target, exists := s.Redirects[path]
	return target, exists
}
