# nanny

Recording log events with all attribute values to for remote inspection through HTTP.


## use as slog Handler

```go
	r := nanny.NewRecorder()
	// fallback only info
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})

	// recorder captures debug too
	l := slog.New(nanny.NewLogHandler(r, h, slog.LevelDebug)) // nanny.LevelTrace
	slog.SetDefault(l)

	//slog.Handle()
	slog.Debug("debug", "c", "d")
	slog.Info("test", "a", "b")
```

## use in HTTP Middleware

```go
	rec := nanny.NewRecorder(nanny.WithMaxEvents(100))

	// record events
	http.Handle("/path", nanny.NewRecordingHTTPHandler(http.HandlerFunc(doPath), rec))
```
Then in your http handle func:

```go
func doPath(w http.ResponseWriter, r *http.Request) {

	rec := nanny.RecorderFromContext(r.Context())
	rec.Group("doPath").
		Record(slog.LevelDebug, "test", "hello").
		Record(slog.LevelInfo, "ev", Bike{Brand: "Trek", Model: "Emonda", Year: "2017"})```

    ...
}
```

## serve the records as JSON

```go
	// serve captured events
	http.Handle("/nanny", nanny.NewBrowser(r))
```


## sample record served as JSON

```json
  {
    "t": "2023-11-08T18:15:14.349402+01:00",
    "l": "DEBUG",
    "n": "bike",
    "g": "checking...",
    "r": "main.Bike",
    "v": {
      "Brand": "Trek",
      "Model": "Emonda",
      "Year": "2017"
    }
```
|field|comment|
|-|-|
|t|timestamp|
|l|log level|
|n|name|
|g|group|
|r|(reflect) type of value|
|v|value