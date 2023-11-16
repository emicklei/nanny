package nanny

import (
	"encoding/json"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strings"

	_ "embed"
)

//go:embed index_template.html
var indexHTML string

const indexHTMLContentType = "text/html; charset=utf-8"

type templateData struct {
	Groups []templateGroup
}
type templateGroup struct {
	Name   string
	Events []Event
}

func (b *Browser) serveIndex(w http.ResponseWriter, r *http.Request) {

	fm := template.FuncMap{
		"valueFormat": func(v any) string {
			d, _ := json.MarshalIndent(v, "", "  ")
			return string(d)
		},
		"levelFormat": func(v any) string {
			switch v.(type) {
			case slog.Level:
				if v.(slog.Level) == LevelTrace {
					return "trace"
				}
				return strings.ToLower(v.(slog.Level).String())
			default:
				return "?"
			}
		},
	}
	tmpl, err := template.New("tt").Funcs(fm).Parse(indexHTML)
	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, err.Error())
		return
	}
	tmplData := buildTemplateData(b.recorder.events)
	w.Header().Set("Content-Type", indexHTMLContentType)
	err = tmpl.Execute(w, tmplData)
	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, err.Error())
		return
	}
}

func buildTemplateData(events []Event) templateData {
	groups := make(map[string][]Event)
	for _, e := range events {
		groups[e.Group] = append(groups[e.Group], e)
	}
	var result []templateGroup
	for k, v := range groups {
		result = append(result, templateGroup{
			Name:   k,
			Events: v,
		})
	}
	return templateData{
		Groups: result,
	}
}
