package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"nhooyr.io/websocket"
)

//go:embed assets/*
var assets embed.FS
var conn *websocket.Conn

func main() {

	assetsRFS, err := fs.Sub(assets, "assets")
	if err != nil {
		log.Fatal(err)
	}
	var assetsFS = http.FS(assetsRFS)

	fs := http.FileServer(assetsFS)
	log.Print("Serving on http://localhost:8888")

	mux := http.NewServeMux()

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ReverseFunc := func(s string) (result string) {
			for _, v := range s {
				result = string(v) + result
			}
			return
		}

		var (
			buffer []byte
			mt     websocket.MessageType
		)

		var err error

		if conn == nil {
			conn, err = websocket.Accept(w, r, &websocket.AcceptOptions{
				CompressionMode: websocket.CompressionDisabled,
				OriginPatterns:  []string{"*"},
			})

			if err != nil {
				panic(err)
			}
		}
		//defer conn.Close(websocket.StatusInternalError, "Websocket: internal Error")
		go func() {
			for {
				//Receive
				mt, buffer, err = conn.Read(context.Background())
				if err != nil {
					if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
						fmt.Println("Connection closed normally")
						conn = nil
						return
					}
					panic(err)
				}
				fmt.Printf("Message received: %s, message type %d\n", string(buffer), mt)

				//Send
				err = conn.Write(
					context.Background(),
					websocket.MessageText,
					[]byte(ReverseFunc(string(buffer))))
				if err != nil {
					if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
						fmt.Println("Connection closed normally")
						conn = nil
						return
					}
					panic(err)
				}

				fmt.Printf("Message send: %s\n", ReverseFunc(string(buffer)))
			}
			//conn.Close(websocket.StatusNormalClosure, "")
		}()
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache")
		if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("content-type", "application/wasm")
		}
		fs.ServeHTTP(w, r)
	})

	err = http.ListenAndServe(":8888", mux)
	if err != nil {
		log.Fatal(err)
	}
}
