VERSION=v0.1.0

all:
	CGO_ENABLED=0 go build -o client --installsuffix cgo client.go
	docker build -t dnsquery:$(VERSION) .	

