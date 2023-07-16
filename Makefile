@all:
	GOOS=js GOARCH=wasm go build -o ./server/assets/realm/main.wasm ./client/main.go
	go build -o realm-client ./client/main.go
	go build -o realm-server ./server/main.go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o realm-server-linux ./server/main.go

clean:
	rm -rf ./server/assets/realm/main.wasm
	rm -rf ./realm-server
	rm -rf ./realm-client
	rm -rf ./realm-server-linux

install_sp:
	ssh sp.crg.eti.br "killall -q -9 realm-server-linux;true"
	scp realm-server-linux sp.crg.eti.br:/home/cesar/
	ssh sp.crg.eti.br "cd /home/cesar && \
		source /home/cesar/env.sh && \
		nohup /home/cesar/realm-server-linux &"

