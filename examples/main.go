package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/emicklei/nanny"
)

type Bike struct {
	Brand, Model, Year string
}

const eventGroupMarker = "httpHandleFunc"

func main() {
	// record max 100 events
	rec := nanny.NewRecorder(
		nanny.WithLogEventGroupOnError(true),
		nanny.WithMaxEventGroups(10),
		nanny.WithGroupMarkers(eventGroupMarker))

	// fallback logger (cannot be the default handler)
	txt := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})

	// handler capturing debug and up
	slog.SetDefault(slog.New(nanny.NewLogHandler(rec, txt, nanny.LevelTrace)))

	// serve do on /do
	http.HandleFunc("/do", do)
	http.HandleFunc("/err", err)

	// serve captured events
	http.Handle("/nanny", nanny.NewBrowser(rec, nanny.WithPageSize(100)))

	// serve
	slog.Info("to create events, open", "url", "http://localhost:8080/do", "port", 8080)
	slog.Info("to browse events, open", "url", "http://localhost:8080/nanny")
	slog.Info("to see events JSON, open", "url", "http://localhost:8080/nanny?do=events")
	slog.Info("to see debug events, open", "url", "http://localhost:8080/nanny?level=debug")
	slog.Info("to see group events, open", "url", "http://localhost:8080/nanny?group=1:do")
	slog.Info("to stop recording, open", "url", "http://localhost:8080/nanny?do=stop")
	slog.Info("to resume recording, open", "url", "http://localhost:8080/nanny?do=resume")
	slog.Info("to flush recorded events, open", "url", "http://localhost:8080/nanny?do=flush")
	slog.Info("to simulate log on error, open", "url", "http://localhost:8080/err")
	http.ListenAndServe(":8080", http.DefaultServeMux)
}

var request_id int64 = 0

func newRequestID() int64 {
	atomic.AddInt64(&request_id, 1)
	return request_id
}

// http handler

func do(w http.ResponseWriter, r *http.Request) {
	// start event group
	glog := slog.Default().With(eventGroupMarker, fmt.Sprintf("%d:do", newRequestID()))

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
	ag.Info("two attributes in attr group in event group", "bike", bike, "color", "red")

	internalDo(glog)

	io.WriteString(w, "done")
}

func internalDo(parentLogger *slog.Logger) {
	// start event group
	glog := parentLogger.With(eventGroupMarker, "internalDo")
	glog.Info("do internal stuff")
}

func err(w http.ResponseWriter, r *http.Request) {
	lg := slog.Default().With(eventGroupMarker, fmt.Sprintf("%d:err", newRequestID()))

	lg.Info("info")
	lg.Debug("debug")
	lg.Log(r.Context(), nanny.LevelTrace, "trace")
	lg.Error("error")
}
