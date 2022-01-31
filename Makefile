build:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/ds-switch ./cmd

clean:
	rm ./bin/ds-switch