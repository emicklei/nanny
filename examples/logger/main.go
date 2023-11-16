package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/emicklei/nanny"
)

type Bike struct {
	Brand, Model, Year string
}

const eventGroupMarker = "httpHandleFunc"

func main() {
	// record max 100 events
	rec := nanny.NewRecorder(nanny.WithMaxEvents(100), nanny.WithGroupMarker(eventGroupMarker))

	// fallback logger (cannot be the default handler)
	txt := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})

	// handler capturing debug and up
	slog.SetDefault(slog.New(nanny.NewLogHandler(rec, txt, slog.LevelDebug)))

	// serve do on /do
	http.HandleFunc("/do", do)

	// serve captured events
	http.Handle("/nanny", nanny.NewBrowser(rec))

	// serve
	slog.Info("to test, open", "url", "http http://localhost:8080/do")
	slog.Info("to see events, open", "url", "http://localhost:8080/nanny")
	http.ListenAndServe(":8080", http.DefaultServeMux)
}

func do(w http.ResponseWriter, r *http.Request) {
	// start event group
	glog := slog.Default().With("do", eventGroupMarker)

	// attributes
	glog.Debug("checking...", slog.Any("bike", Bike{Brand: "Trek", Model: "Emonda", Year: "2017"}))

	// wont see this event in the recorder
	glog.Info("no attributes")

	// attributes without group
	glog.Info("two attributes", slog.String("bike", "Specialized"), slog.String("size", "29inch"))

	// attribute group
	ag := glog.WithGroup("myattrs")
	ag.Info("two attributes in group", slog.String("bike", "Trek"), slog.String("size", "27inch"))

	io.WriteString(w, "done")
}
