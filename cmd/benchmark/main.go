package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"strconv"
	"time"
	"wsChatroom/internal"
)

var (
	userNum       int
	loginInterval time.Duration
	msgInterval   time.Duration
)

func init() {
	flag.IntVar(&userNum, "u", 500, "登录用户数")
	flag.DurationVar(&loginInterval, "l", 5e9, "用户陆续登录时间间隔")
	flag.DurationVar(&msgInterval, "m", 1*time.Minute, "用户发送消息时间间隔")
}

func main() {
	flag.Parse()

	for i := 0; i < userNum; i++ {
		go UserConnect("user" + strconv.Itoa(i))
		time.Sleep(loginInterval)
	}
	select {}
}

func UserConnect(nickname string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, "ws://127.0.0.1:2022/ws?nickname="+nickname, nil)
	if err != nil {
		log.Println("dial error:", err)
		return
	}
	defer conn.Close(websocket.StatusInternalError, "内部错误！")
	go sendMessage(conn, nickname)
	ctx = context.Background()

	for {
		var message internal.Message
		err = wsjson.Read(ctx, conn, &message)
		if err != nil {
			log.Println("receive msg error:", err)
			continue
		}
		if message.ClientSendTime.IsZero() {
			continue
		}
		if d := time.Now().Sub(message.ClientSendTime); d > 1*time.Second {
			fmt.Printf("接收到服务端响应(%d)：%#v\n", d.Milliseconds(), message)
		}
	}
	conn.Close(websocket.StatusNormalClosure, "")
}

func sendMessage(conn *websocket.Conn, nickname string) {
	ctx := context.Background()
	i := 1
	for {
		msg := map[string]string{
			"content":   "来自" + nickname + "的消息:" + strconv.Itoa(i),
			"send_time": strconv.FormatInt(time.Now().UnixNano(), 10),
		}
		err := wsjson.Write(ctx, conn, msg)
		if err != nil {
			log.Println("send msg error:", err, "nickname:", nickname, "no:", i)
		}
		i++

		time.Sleep(msgInterval)
	}
}
