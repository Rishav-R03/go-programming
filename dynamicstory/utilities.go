package dynamicstory

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
)

type Option struct {
	Text string `json:"text"`
	Arc  string `json:"arc"`
}

type Chapter struct {
	Title   string   `json:"title"`
	Story   []string `json:"story"`
	Options []Option `json:"options"`
}

// Story maps story arc keys (eg "intro") to Chapter objects.
type Story map[string]Chapter

func JsonStory(r io.Reader) (Story, error) {
	d := json.NewDecoder(r)
	var story Story
	if err := d.Decode(&story); err != nil {
		return nil, err
	}
	return story, nil
}

type HandlerOption func(*handler)

func WithTemplate(t *template.Template) HandlerOption {
	return func(h *handler) {
		h.t = t
	}
}

func WithPathFunc(fn func(*http.Request) string) HandlerOption {
	return func(h *handler) {
		h.pathFn = fn
	}
}

var defaultStoryTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <title>Choose Your Own Adventure</title>
    <style>
      body {
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
        background-color: #f7fafc;
        color: #2d3748;
        max-width: 700px;
        margin: 40px auto;
        padding: 20px;
      }
      .card {
        background: #ffffff;
        padding: 30px;
        border-radius: 8px;
        box-shadow: 0 4px 6px rgba(0,0,0,0.1);
      }
      h1 { color: #1a202c; border-bottom: 2px solid #e2e8f0; padding-bottom: 10px; }
      p { line-height: 1.6; margin-bottom: 1.2em; }
      ul { list-style: none; padding: 0; }
      li { margin-bottom: 10px; }
      a {
        display: block;
        padding: 12px 16px;
        background: #3182ce;
        color: white;
        text-decoration: none;
        border-radius: 4px;
        transition: background 0.2s;
      }
      a:hover { background: #2b6cb0; }
    </style>
  </head>
  <body>
    <section class="card">
      <h1>{{.Title}}</h1>
      {{range .Story}}
        <p>{{.}}</p>
      {{end}}

      {{if .Options}}
        <ul>
          {{range .Options}}
            <li><a href="/{{.Arc}}">{{.Text}}</a></li>
          {{end}}
        </ul>
      {{else}}
        <p><strong>The End</strong></p>
        <p><a href="/intro">Restart Adventure</a></p>
      {{end}}
    </section>
  </body>
</html>
`

var defaultTpl = template.Must(template.New("").Parse(defaultStoryTemplate))

type handler struct {
	s      Story
	t      *template.Template
	pathFn func(*http.Request) string
}

func defaultPathFn(r *http.Request) string {
	path := strings.TrimSpace(r.URL.Path)
	if path == "" || path == "/" {
		path = "/intro"
	}
	return path[1:]
}
func NewHandler(s Story, opts ...HandlerOption) http.Handler {
	h := handler{
		s:      s,
		t:      defaultTpl,
		pathFn: defaultPathFn,
	}
	for _, opt := range opts {
		opt(&h)
	}
	return h
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := h.pathFn(r)
	if chapter, ok := h.s[path]; ok {
		err := h.t.Execute(w, chapter)
		if err != nil {
			log.Printf("%v", err)
			http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "Chapter not found.", http.StatusNotFound)
}
