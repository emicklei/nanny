package nanny

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	_ "embed"
)

const indexHTMLContentType = "text/html; charset=utf-8"

//go:embed index.html
var contentHTML string

func (b *Browser) serveStaticIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", indexHTMLContentType)
	// serve content
	if os.Getenv("DEV") != "" {
		fmt.Println("reloading ../index.html...")
		content, err := os.ReadFile("../index.html")
		if err != nil {
			w.WriteHeader(500)
			io.WriteString(w, err.Error())
			return
		}
		contentHTML = string(content)
	}
	// apply HTML insertions if any
	if b.options.EndHTMLHeadFunc != nil {
		endHTML := b.options.EndHTMLHeadFunc()
		contentHTML = strings.Replace(contentHTML, "<!--EndHTMLHeadFunc-->", endHTML, 1)
	}
	if b.options.BeforeHTMLTableFunc != nil {
		beforeHTML := b.options.BeforeHTMLTableFunc()
		contentHTML = strings.Replace(contentHTML, "<!--BeforeHTMLTableFunc-->", beforeHTML, 1)
	}
	if b.options.AfterHTMLFiltersFunc != nil {
		afterHTML := b.options.AfterHTMLFiltersFunc()
		contentHTML = strings.Replace(contentHTML, "<!--AfterHTMLFiltersFunc-->", afterHTML, 1)
	}
	io.WriteString(w, contentHTML)
}
