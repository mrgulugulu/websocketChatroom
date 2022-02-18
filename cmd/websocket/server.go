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
		fmt.Fprintln(w, "HTTP, HELLO")
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, req *http.Request) {
		conn, err := websocket.Accept(w, req, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close(websocket.StatusInternalError, "内部出错了!")

		// 超时就cancel
		ctx, cancel := context.WithTimeout(req.Context(), time.Second*10)
		defer cancel()

		var v interface{}
		err = wsjson.Read(ctx, conn, &v) //
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("接受到客户端：%v\n", v)

		err = wsjson.Write(ctx, conn, "hello websocket client")
		if err != nil {
			log.Println(err)
			return
		}

		conn.Close(websocket.StatusNormalClosure, "")
	})
	log.Fatal(http.ListenAndServe(":2021", nil))
}
