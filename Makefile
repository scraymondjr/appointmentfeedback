.PHONY: build clean deploy test integration

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/api cmd/api/main.go

clean:
	rm -rf ./bin ./vendor

deploy: clean build
	sls deploy --verbose

test:
	go test ./...

integration:
	go test -tags=integration ./...