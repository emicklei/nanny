# nanny

[![GoDoc](https://pkg.go.dev/badge/github.com/emicklei/nanny)](https://pkg.go.dev/github.com/emicklei/nanny)

Records a sliding window of slog events with all attribute values for remote inspection through HTTP.


## usage

```go
	import "github.com/emicklei/nanny"

	...

	nanny.SetupDefault()
	slog.Debug("debug", "hello", "world")
```

or by composing the setup yourself:


```go
	r := nanny.NewRecorder(
		nanny.WithLogEventGroupOnError(true),
		nanny.WithMaxEventGroups(100),
		// any of these attribute keys can be used for event grouping
		nanny.WithGroupMarkers(
			"func",
			"X-Cloud-Trace-Context",
			"X-Request-ID"))

	// recorder captures debug 
	l := slog.New(nanny.NewLogHandler(r, slog.Default().Handler(), slog.LevelDebug)) // or nanny.LevelTrace
	
	// replace the default logger
	slog.SetDefault(l)

	// serve the events, protect using basic auth
	http.Handle("/nanny", nanny.NewBasicAuthHandler(
		nanny.NewBrowser(rec),
		os.Getenv("NANNY_USER"),
		os.Getenv("NANNY_PASSWORD")))

	slog.Debug("debug", "hello", "world")
```

Then after starting your HTTP service, you can access `/nanny` to see and explore your log events.
 

## event groups

Events can be grouped e.g. by function name or for the processing of a specific HTTP request.

```go
	l := slog.Default().With("func", "myFunctionName")
	l.Debug("var", "key", "value")
```

Here `func` is the default event group marker.
You can change the group keys to whatever you want using the RecorderOption `WithGroupMarkers`.

## log event group on error

With this option, if an Error event is recorded then all leading debug and trace events in the same group are logger first.

```go
	r := nanny.NewRecorder(nanny.WithLogEventGroupOnError(true))
	...
```

## sample record served as JSON

```json
  {
    "t": "2023-11-08T18:15:14.349402+01:00",
    "l": "DEBUG",
    "g" : "some group", 
    "m": "checking...", 
    "a": {
      "bike": {
		"Brand": "Trek",
      	"Model": "Emonda",
      	"Year": "2017"
	  }
    }
```
|field|comment|
|-|-|
|t|timestamp|
|l|log level|
|g|group|
|m|message|  
|a|attributes|
