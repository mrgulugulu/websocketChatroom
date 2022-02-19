package server

import (
	"net/http"
	"wsChatroom/internal"
)

func RegisterHandle() {
	// 没返回结果的一般都用go
	go internal.Broadcaster.Start()

	http.HandleFunc("/", homeHandleFunc)
	http.HandleFunc("/user_list", userListHandleFunc)
	http.HandleFunc("/ws", WebSocketHandleFunc)
}
