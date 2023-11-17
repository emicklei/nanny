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
	slog.Info("to test, open", "url", "http://localhost:8080/do", "port", 8080)
	slog.Info("to see events, open", "url", "http://localhost:8080/nanny")
	http.ListenAndServe(":8080", http.DefaultServeMux)
}

func do(w http.ResponseWriter, r *http.Request) {
	// start event group
	glog := slog.Default().With("do", eventGroupMarker)

	// attributes
	bike := Bike{Brand: "Trek", Model: "Emonda", Year: "2017"}
	glog.Debug("checking...", slog.Any("bike", bike))

	// wont see this event in the recorder
	glog.Info("no attributes")

	// attributes without group
	glog.Info("two attributes", slog.String("bike", "Specialized"), slog.String("size", "29inch"))

	// attribute group within event group
	ag := glog.WithGroup("myattrs")

	// modify it
	bike.Brand = "Specialized"
	bike.Year = "2018"

	ag.Info("two attributes in attr group in event group", "bike", bike)

	io.WriteString(w, "done")
}
