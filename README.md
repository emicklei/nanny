# nanny

Recording log events with all attribute values to for remote inspection through HTTP.


## usage

```go
	r := nanny.NewRecorder()
	// fallback only info
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})

	// recorder captures debug too
	l := slog.New(nanny.NewLogHandler(r, h, slog.LevelDebug)) // nanny.LevelTrace
	slog.SetDefault(l)

	slog.Debug("debug", "c", "d")
	slog.Info("test", "a", "b")
```
 

## event groups

Events can be grouped e.g. by function name or for the processing of a specific HTTP request.

```go
	l := slog.Default().With("func", "myFunctionName")
	l.Debug("var", "value")	
```

Here `func` is the default event group marker.
You can change this value to whatever you want using the RecorderOption `WithGroupMarker`.


## serve the records

```go
	// serve captured events
	http.Handle("/nanny", nanny.NewBrowser(r))
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