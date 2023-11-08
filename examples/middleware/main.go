package main

import (
	"net/http"

	"github.com/emicklei/nanny"
)

type Bike struct {
	Brand, Model, Year string
}

func main() {
	rec := nanny.NewRecorder(nanny.WithMaxEvents(100))

	// record events
	http.Handle("/do", nanny.NewRecordingHTTPHandler(http.HandlerFunc(do), rec))

	// serve captured events
	http.Handle("/nanny", nanny.NewBrowser(rec))

	// serve
	http.ListenAndServe(":8080", http.DefaultServeMux)
}

func do(w http.ResponseWriter, r *http.Request) {

	nanny.RecorderFromContext(r.Context()).Group("some operation").
		Record("test", "hello").
		Record("ev", Bike{Brand: "Trek", Model: "Emonda", Year: "2017"})

	w.Write([]byte("hello nanny"))
}
