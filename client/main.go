package main

import (
	"context"
	"fmt"
	_ "image/png"
	"log"
	"strings"

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

const MyMessage = "test test test"

var conn *websocket.Conn
var i int

func (g *Game) Update() error {
	g.keys = inpututil.AppendPressedKeys(g.keys[:0])

	var (
		buffer []byte
		mt     websocket.MessageType
	)

	var err error

	if conn == nil {
		log.Println("connecting...")
		conn, _, err = websocket.Dial(context.Background(), "ws://127.0.0.1:8888/ws", nil)
		if err != nil {
			panic(err)
		}
	}
	//defer conn.Close(websocket.StatusInternalError, "Websocket: internal Error")

	//Send
	conn.Write(context.Background(), websocket.MessageBinary, []byte(MyMessage))
	fmt.Printf("Message send: %s\n", MyMessage)

	//Receive
	mt, buffer, err = conn.Read(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Message received: %s, message type %d\n", string(buffer), mt)
	i++
	if i == 10 {
		conn.Close(websocket.StatusNormalClosure, "Websocket: normal closure")
		conn = nil
		log.Println("Connection closed")
		i = 0
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	keyStrs := []string{}
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
