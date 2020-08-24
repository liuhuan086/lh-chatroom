package main

import (
	"context"
	"fmt"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "ws://localhost:2021/ws", nil)
	if err != nil {
		panic(err)
	}

	defer c.Close(websocket.StatusInternalError, "Internet Error.")

	err = wsjson.Write(ctx, c, "Hello WebSocket Server.")
	if err != nil {
		panic(err)
	}

	var v interface{}

	err = wsjson.Read(ctx, c, &v)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Accept Server Endpoint response: %v\n", v)
	_ = c.Close(websocket.StatusNormalClosure, "")
}
