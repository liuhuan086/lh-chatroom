package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintln(w, "HTTP, Hello")
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, req *http.Request) {
		conn, err := websocket.Accept(w, req, nil)
		if err != nil {
			log.Println(err)
			return
		}

		defer conn.Close(websocket.StatusInternalError, "Internet Server Error!")

		ctx, cancel := context.WithTimeout(req.Context(), time.Second*10)
		defer cancel()

		var v interface{}
		err = wsjson.Read(ctx, conn, &v)
		if err != nil {
			log.Printf("Accept client: %v\n", v)
			return
		}

		err = wsjson.Write(ctx, conn, "Hello WebSocket Client!")
		if err != nil {
			log.Printf("Send WebSocket Error, error is %v", err)
			return
		}

		_ = conn.Close(websocket.StatusNormalClosure, "")

	})

	log.Fatal(http.ListenAndServe(":2021", nil))
}
