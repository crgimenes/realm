package main

import (
	"context"
	"crypto/rand"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

type connectedUser struct {
	conn      *websocket.Conn
	nick      string
	sessionID string
	x, y      int
}

//go:embed assets/*
var assets embed.FS

var (
	mutex          sync.Mutex
	connectedUsers = make(map[string]connectedUser)
)

func removeUser(sessionID string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(connectedUsers, sessionID)
}

func addUser(sessionID string, user connectedUser) {
	mutex.Lock()
	defer mutex.Unlock()
	connectedUsers[sessionID] = user
}

func broadcast(buffer []byte) {
	mutex.Lock()
	defer mutex.Unlock()
	for _, user := range connectedUsers {
		err := send(user.conn, buffer)
		if err != nil {
			log.Println(err)
			removeUser(user.sessionID)
		}
	}
}

func send(conn *websocket.Conn, buffer []byte) error {
	err := conn.Write(
		context.Background(),
		websocket.MessageText,
		buffer,
	)
	if err != nil {
		if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
			fmt.Println("Connection closed normally")
			conn = nil
			return nil
		}
		return err
	}

	fmt.Printf("Message send: %s\n", string(buffer))
	return nil
}

func parseMessage(userID string, conn *websocket.Conn, buffer []byte) error {
	log.Printf("Parsing message: %s\n", string(buffer))
	return nil
}

func randomID() string {
	const (
		length  = 16
		charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	)
	lenCharset := byte(len(charset))
	b := make([]byte, length)
	rand.Read(b)
	for i := 0; i < length; i++ {
		b[i] = charset[b[i]%lenCharset]
	}
	return string(b)
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	var (
		conn   *websocket.Conn
		buffer []byte
		mt     websocket.MessageType
		err    error
	)

	conn, err = websocket.Accept(w, r, &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionDisabled,
		OriginPatterns:  []string{"*"},
	})
	if err != nil {
		log.Println(err)
		return
	}

	user := connectedUser{
		conn:      conn,
		nick:      "anonymous",
		x:         0,
		y:         0,
		sessionID: randomID(),
	}

	addUser(user.sessionID, user)

	initTime := time.Now()

	go func() {
		for {
			//Receive
			initTime = time.Now()
			mt, buffer, err = conn.Read(context.Background())
			if err != nil {
				conn = nil
				removeUser(user.sessionID)
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
					fmt.Println("Connection closed normally")
					return
				}
				log.Println(err)
				return
			}
			timeElapsed := time.Since(initTime)
			fmt.Printf("Time elapsed: %s\n", timeElapsed)
			fmt.Printf("Message received: %s, message type %d\n", string(buffer), mt)

			//Parse
			err = parseMessage(user.sessionID, conn, buffer)
			if err != nil {
				log.Println(err)
				removeUser(user.sessionID)
				return
			}
		}
	}()
}

func main() {

	assetsRFS, _ := fs.Sub(assets, "assets")
	var assetsFS = http.FS(assetsRFS)

	fs := http.FileServer(assetsFS)
	log.Print("Serving on http://localhost:8888")

	mux := http.NewServeMux()

	mux.HandleFunc("/ws", websocketHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache")
		if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("content-type", "application/wasm")
		}
		fs.ServeHTTP(w, r)
	})

	err := http.ListenAndServe(":8888", mux)
	if err != nil {
		log.Fatal(err)
	}
}
