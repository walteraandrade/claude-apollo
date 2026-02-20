BINARY=apollo

.PHONY: build run test clean install

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY)

test:
	go test -race ./...

clean:
	rm -f $(BINARY)

install: build
	cp $(BINARY) $(GOPATH)/bin/$(BINARY) 2>/dev/null || cp $(BINARY) ~/go/bin/$(BINARY)
