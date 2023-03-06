.PHONY: docker clean

build: $(shell find . -iname '*.go')
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o bin/client-linux main.go
	
clean:
	rm -rf bin
