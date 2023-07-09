@all:
	GOOS=js GOARCH=wasm go build -o ./server/assets/main.wasm ./client/main.go
	go build -o realm-client ./client/main.go
	go build -o realm-server ./server/main.go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o realm-server-linux ./server/main.go

clean:
	rm -rf ./server/assets/main.wasm
	rm -rf ./realm-server
	rm -rf ./realm-client
	rm -rf ./realm-server-linux



