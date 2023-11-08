package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/emicklei/nanny"
)

type Bike struct {
	Brand, Model, Year string
}

func main() {
	rec := nanny.NewRecorder(nanny.WithMaxEvents(100))
	reclog := slog.New(nanny.NewLogHandler(rec, slog.Default().Handler()))
	slog.SetDefault(reclog)

	http.HandleFunc("/do", do)

	// serve captured events
	http.Handle("/nanny", nanny.NewBrowser(rec))

	// serve
	http.ListenAndServe(":8080", http.DefaultServeMux)
}

func do(w http.ResponseWriter, r *http.Request) {
	fmt.Println("do")
	slog.Info("message", slog.Any("bike", Bike{Brand: "Trek", Model: "Emonda", Year: "2017"}))
	io.WriteString(w, "done")
}
