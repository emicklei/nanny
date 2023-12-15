package nanny

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	_ "embed"
)

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

//go:embed index.html
var contentHTML []byte

func (b *Browser) serveStaticIndex(w http.ResponseWriter, r *http.Request) {
	if os.Getenv("DEV") != "" {
		fmt.Println("reloading ../index.html...")
		content, err := os.ReadFile("../index.html")
		if err != nil {
			w.WriteHeader(500)
			io.WriteString(w, err.Error())
			return
		}
		contentHTML = content
	}
	w.Header().Set("Content-Type", indexHTMLContentType)
	w.Write(contentHTML)
}
