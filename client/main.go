package main

import (
	"context"
	"fmt"
	_ "image/png"
	"log"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"nhooyr.io/websocket"
)

const (
	screenWidth  = 320
	screenHeight = 240
)

type Game struct {
	keys []ebiten.Key
}

var (
	conn   *websocket.Conn
	tosend = make(chan []byte, 100)
)

func parseMessage(buffer []byte) error {
	log.Printf("Parse message received: %s\n", buffer)
	return nil
}

func receiveLoop() {
	var (
		buffer []byte
		mt     websocket.MessageType
		err    error
	)

	for {
		if conn == nil {
			return
		}
		//Receive
		mt, buffer, err = conn.Read(context.Background())
		if err != nil {
			conn = nil
			log.Println(err)
			return
		}
		log.Printf("Message received: %s, message type %d\n", string(buffer), mt)

		//conn.Close(websocket.StatusNormalClosure, "Websocket: normal closure")

		//Parse
		err = parseMessage(buffer)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func directSend(msg []byte) error {
	//Send

	if conn == nil {
		return fmt.Errorf("conn is nil")
	}

	err := conn.Write(context.Background(), websocket.MessageBinary, msg)
	if err != nil {
		log.Println(err)
		conn = nil
		return err
	}
	log.Printf("Message send: %s\n", string(msg))
	return nil
}

func send(msg []byte) {
	tosend <- msg
}

func sendLoop() {
	var err error
	for {
		select {
		case <-time.After(1 * time.Second):
			err = directSend([]byte("ping"))
			if err != nil {
				log.Println(err)
				return
			}
		case msg := <-tosend:
			err = directSend(msg)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func (g *Game) Update() error {
	g.keys = inpututil.AppendPressedKeys(g.keys[:0])

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		log.Println(g.keys)
	}

	var err error

	if conn == nil {
		log.Println("connecting...")
		//conn, _, err = websocket.Dial(context.Background(), "ws://127.0.0.1:8888/ws", nil)
		conn, _, err = websocket.Dial(context.Background(), "wss://sp.crg.eti.br/ws", nil)
		if err != nil {
			conn = nil
			log.Println(err)
			return nil
		}

		go receiveLoop()
		go sendLoop()
	}
	//defer conn.Close(websocket.StatusInternalError, "Websocket: internal Error")

	return nil
}

var keyStrs = []string{}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, p := range g.keys {
		keyStrs = append(keyStrs, p.String())
	}
	ebitenutil.DebugPrint(screen, strings.Join(keyStrs, ", "))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("realm")
	ebiten.SetRunnableOnUnfocused(true)
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
