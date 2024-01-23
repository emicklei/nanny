package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/emicklei/nanny"
)

type Bike struct {
	Brand, Model, Year string
}

const eventGroupMarker = "httpHandleFunc"

var N = flag.Int("N", 10, "number of events to create")

func main() {
	flag.Parse()

	// create recorder
	// record max 100 events
	rec := nanny.NewRecorder(nanny.RecorderOptions{
		GroupMarkers:   []string{eventGroupMarker},
		MaxEventGroups: 10,
	})

	// fallback logger (cannot be the default handler)
	txt := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})

	// handler capturing debug and up
	slog.SetDefault(slog.New(nanny.NewLogHandler(rec, txt, nanny.LevelTrace)))

	// serve do on /do
	http.HandleFunc("/do", do)
	http.HandleFunc("/err", err)
	http.HandleFunc("/hidden", hidden)

	// serve captured events
	http.Handle("/nanny", nanny.NewBrowser(rec, nanny.BrowserOptions{PageSize: 10}))

	slog.Info("generating events...", "N", *N)

	// create events
	for i := 0; i < *N; i++ {
		handleDo()
	}

	// serve
	slog.Info("to create events, open", "url", "http://localhost:8080/do", "port", 8080)
	slog.Info("to browse events, open", "url", "http://localhost:8080/nanny")
	slog.Info("to see events JSON, open", "url", "http://localhost:8080/nanny?do=events")
	slog.Info("to see debug events, open", "url", "http://localhost:8080/nanny?level=debug")
	slog.Info("to see group events, open", "url", "http://localhost:8080/nanny?group=1:do")
	slog.Info("to stop recording, open", "url", "http://localhost:8080/nanny?do=stop")
	slog.Info("to resume recording, open", "url", "http://localhost:8080/nanny?do=resume")
	slog.Info("to flush recorded events, open", "url", "http://localhost:8080/nanny?do=flush")
	slog.Info("to simulate log with debug and trace, open", "url", "http://localhost:8080/hidden")
	slog.Info("to simulate log on error, open", "url", "http://localhost:8080/err")

	// start http server
	http.ListenAndServe(":8080", http.DefaultServeMux)
}

var request_id int64 = 0

func newRequestID() int64 {
	atomic.AddInt64(&request_id, 1)
	return request_id
}

// http handler

func do(w http.ResponseWriter, r *http.Request) {
	handleDo()
	io.WriteString(w, "done")
}

func handleDo() {
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

	glog.Info("info")
	glog.Debug("debug")
	nanny.Trace(glog, "trace")
	glog.Warn("warn")
	glog.Error("error", "err", errors.New("some error"))
}

func internalDo(parentLogger *slog.Logger) {
	// start event group
	glog := parentLogger.With(eventGroupMarker, "internalDo")
	glog.Info("do internal stuff")
	glog.Info(`
	Lorem ipsum is a placeholder text commonly used.
	`, "content", `
	“Crocubot. So, you’re a cold, unfeeling reptile and also an equally cold, and unfeeling machine? Yes. So your origin is what? You fell in a vat of redundancy?”
	`, "pi", math.Pi, "sqrte", math.SqrtE, "ln2", math.Ln2)
}

func err(w http.ResponseWriter, r *http.Request) {
	lg := slog.Default().With(eventGroupMarker, fmt.Sprintf("%d:err", newRequestID()))

	lg.Info("any debug or trace events that are recorded for this group will be shown in the output when an error occurs")
	lg.Debug("debug")
	nanny.Trace(lg, "trace")
	lg.Warn("warn")
	lg.Error("error")
}

func hidden(w http.ResponseWriter, r *http.Request) {
	lg := slog.Default().With(eventGroupMarker, fmt.Sprintf("%d:hidden", newRequestID()))

	lg.Info("debug or trace events are recorded but not shown in the output")
	lg.Debug("debug")
	nanny.Trace(lg, "trace")
}
