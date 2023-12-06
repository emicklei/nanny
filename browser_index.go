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
	"strconv"
	"strings"
	"time"

	_ "embed"
)

//go:embed index_template.html
var indexHTML string

const indexHTMLContentType = "text/html; charset=utf-8"

type templateData struct {
	Events         []Event
	EventsSeen     int64
	Since          time.Time
	OffsetPrevious int
	OffsetNext     int
}

type eventFilter struct {
	level  string
	group  string
	offset int
	count  int
}

func (b *Browser) serveIndex(w http.ResponseWriter, r *http.Request) {
	list := b.recorder.snapshotEvents()
	filter := eventFilter{
		level: r.Form.Get("level"),
		group: r.Form.Get("group"),
	}
	if count := r.Form.Get("count"); count != "" {
		filter.count, _ = strconv.Atoi(count)
	} else {
		// fallback to browser option
		filter.count = b.pageSize
	}
	if offset := r.Form.Get("offset"); offset != "" {
		filter.offset, _ = strconv.Atoi(offset)
	} else {
		// last page first
		filter.offset = len(list) - filter.count
	}
	// make sure offset is in range
	if filter.offset < 0 {
		filter.offset = 0
	}
	if filter.offset > len(list) {
		filter.offset = len(list)
	}
	fm := template.FuncMap{
		"timeFormat": func(v any) string {
			return v.(time.Time).Format("2006-01-02 15:04:05.000")
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
				l := v.(slog.Level)
				switch l {
				case slog.LevelDebug:
					return "debug"
				default:
					return strings.ToLower(v.(slog.Level).String())
				}
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
	tmplData := templateData{
		Events:         filtered(list, filter),
		EventsSeen:     b.recorder.stats.Count,
		Since:          b.recorder.stats.Started,
		OffsetPrevious: filter.offset - filter.count,
		OffsetNext:     filter.offset + filter.count,
	}
	w.Header().Set("Content-Type", indexHTMLContentType)
	err = tmpl.Execute(w, tmplData)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
}

func filtered(events []Event, filter eventFilter) (list []Event) {
	// if group filter active then ignore paging
	if filter.group != "" {
		filter.offset = 0
		filter.count = len(events)
	}
	for i := filter.offset; i < len(events) && len(list) < filter.count; i++ {
		each := events[i]
		if filter.level != "" && strings.ToLower(each.Level.String()) != filter.level {
			continue
		}
		if filter.group != "" {
			if each.Group == "" {
				continue
			}
			if !strings.HasPrefix(each.Group, filter.group) {
				continue
			}
		}
		list = append(list, each)
	}
	return
}

func shortValueFormat(v any) string {
	if v == nil {
		return ""
	}
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
		if v == nil {
			fmt.Fprintf(sb, "%s=nil ", k)
			continue
		}
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
