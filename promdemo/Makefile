VERSION=v0.1.0

all:
	CGO_ENABLED=0 go build -o demo --installsuffix cgo demo.go
	docker build -t promdemo:$(VERSION) .

