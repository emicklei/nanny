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

	//slog.Handle()
	slog.Debug("debug", "c", "d")
	slog.Info("test", "a", "b")
```
 

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
	"g" : "corr-id",
    "n": "bike",
    "m": "checking...",
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
|g|group|
|m|message|
|n|name|
|r|(reflect) type of value|
|v|value