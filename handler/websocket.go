package handler

import (
	"context"
	"log"
	"net/http"
	"realm/util"
	"sync"
	"time"

	"nhooyr.io/websocket"

	"realm/globalconst"
)

type connectedUser struct {
	conn      *websocket.Conn
	nick      string
	sessionID string
	x, y      int
}

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
			log.Println("Connection closed normally")
			conn = nil
			return nil
		}
		return err
	}

	log.Printf("Message send: %s\n", string(buffer))
	return nil
}

func parseMessage(userID string, conn *websocket.Conn, buffer []byte) error {
	if len(buffer) == 0 {
		return nil
	}

	p := buffer[0]

	switch p {
	case '!':
		log.Printf("Ping received from %s, sessionID: %s\n",
			userID,
			connectedUsers[userID].sessionID)
	case '~':
		buffer[0] = '.' //Replace ~ with .
		for _, user := range connectedUsers {
			if user.sessionID == userID {
				continue
			}
			err := send(user.conn, buffer)
			if err != nil {
				log.Println(err)
				removeUser(user.sessionID)
			}
		}
	case '.':
		log.Printf("Message received: %s\n", string(buffer))
	default:
		log.Printf("Unknown message received: %s\n", string(buffer))
	}

	return nil
}

func Websocket(w http.ResponseWriter, r *http.Request) {
	var (
		conn   *websocket.Conn
		buffer []byte
		mt     websocket.MessageType
		err    error
	)

	// get cookie
	cookie, err := r.Cookie(globalconst.CookieName)
	if err != nil {
		log.Printf("Cookie: %v\n", err)
		return
	}

	// get session
	sessionID := cookie.Value
	if sessionID == "" {
		log.Println("SessionID is empty")
		sessionID = util.RandomID()
	}

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
		sessionID: sessionID,
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
					log.Println("Connection closed normally")
					return
				}
				log.Println(err)
				return
			}
			timeElapsed := time.Since(initTime)
			log.Printf("Time elapsed: %s\n", timeElapsed)
			log.Printf("Message received: %s, message type %d\n", string(buffer), mt)

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
