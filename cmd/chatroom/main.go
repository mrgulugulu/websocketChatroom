package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"wsChatroom/global"
	"wsChatroom/server"
)

// 参考自https://github.com/go-programming-tour-book/chatroom
var (
	addr   = ":2022"
	banner = `
    ____              _____
   |    |    |   /\     |
   |    |____|  /  \    | 
   |    |    | /----\   |
   |____|    |/      \  |
miniChatRoom，start on：%s
`
)

// 在main运行前会运行一次
func init() {
	global.Init()
}

func main() {
	fmt.Printf(banner+"\n", addr)
	server.RegisterHandle()
	log.Fatal(http.ListenAndServe(addr, nil))
}
