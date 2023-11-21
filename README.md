# nanny

[![GoDoc](https://pkg.go.dev/badge/github.com/emicklei/nanny)](https://pkg.go.dev/github.com/emicklei/nanny)

Records a sliding window of slog events with all attribute values for remote inspection through HTTP.


## usage

```go
	nanny.SetupDefault()
	slog.Debug("debug", "hello", "world")
```

or by composing the setup yourself:


```go
	r := nanny.NewRecorder()

	// recorder captures debug 
	l := slog.New(nanny.NewLogHandler(r, slog.Default().Handler(), slog.LevelDebug)) // or nanny.LevelTrace
	
	// replace the default logger
	slog.SetDefault(l)

	// serve the events
	http.Handle("/nanny", nanny.NewBrowser(r))

	slog.Debug("debug", "hello", "world")
```
 

## event groups

Events can be grouped e.g. by function name or for the processing of a specific HTTP request.

```go
	l := slog.Default().With("func", "myFunctionName")
	l.Debug("var", "value")	
```

Here `func` is the default event group marker.
You can change the group keys to whatever you want using the RecorderOption `WithGroupMarkers`.


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