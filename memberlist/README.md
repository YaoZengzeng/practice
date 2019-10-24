### Start First Node

```
$ go run main.go
```

### Start Second Node

```
$ go run main.go --port 4002 --bind-port 8002 --members 127.0.0.1:8001
```

### Test

```
$ curl "http://127.0.0.1:4001/add?key=foo&val=bar"

$ curl "http://127.0.0.1:4002/get?key=foo"
bar
```

