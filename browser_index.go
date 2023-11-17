package nanny

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

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
		"timeFormat": func(v any) string {
			s := v.(time.Time).Format("2006-01-02 15:04:05.99")
			// bug in sdk?
			if len(s) != 22 {
				s += "0"
			}
			return s
		},
		"valueFormat": func(v any) string {
			d, _ := json.MarshalIndent(v, "", "  ")
			return string(d)
		},
		"shortValueFormat": shortValueFormat,
		"levelFormat": func(v any) string {
			switch v.(type) {
			case slog.Level:
				if v.(slog.Level) == LevelTrace {
					return "trace"
				}
				return strings.ToLower(v.(slog.Level).String())
			default:
				return "note"
			}
		},
		"keysFormat": func(v any) string {
			return strings.Join(v.([]string), ",")
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

func shortValueFormat(v any) string {
	m, ok := v.(map[string]any)
	if !ok {
		d, _ := json.Marshal(v)
		return string(d)
	}
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sb := new(strings.Builder)
	for _, k := range keys {
		v := m[k]
		if reflect.TypeOf(v).Kind() == reflect.Map {
			fmt.Fprintf(sb, "%s={...} ", k)
			continue
		}
		s, ok := v.(string)
		if ok {
			fmt.Fprintf(sb, "%s=%q ", k, s)
		} else {
			fmt.Fprintf(sb, "%s=%v ", k, v)
		}
	}
	return sb.String()
}
