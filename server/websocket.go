package server

import (
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"wsChatroom/internal"
)

func WebSocketHandleFunc(w http.ResponseWriter, req *http.Request) {
	conn, err := websocket.Accept(w, req, nil)
	if err != nil {
		log.Println("websocket accept error:", err)
		return
	}

	// 构建新用户实例
	token := req.FormValue("token")
	nickname := req.FormValue("nickname")
	if l := len(nickname); l < 2 || l > 20 {
		log.Println("nickname illegal: ", nickname)
		wsjson.Write(req.Context(), conn, internal.NewErrorMessage("非法昵称，昵称长度4-20"))
		conn.Close(websocket.StatusUnsupportedData, "nickname illegal!")
		return
	}
	if !internal.Broadcaster.CanEnterRoom(nickname) {
		log.Println("昵称已经存在:", nickname)
		wsjson.Write(req.Context(), conn, internal.NewErrorMessage("该昵称已经存在"))
		conn.Close(websocket.StatusUnsupportedData, "nickname exists!")
		return
	}

	// 如果nickname已存在则连接会断开，否则会创建该用户的实例
	userHasToken := internal.NewUser(conn, token, nickname, req.RemoteAddr)

	// 开启给单独用户发送信息的goroutine，这是个长期运行的goroutine，存在泄露的风险，记得关闭
	go userHasToken.SendMessage(req.Context())

	// 给用户发送欢迎消息
	userHasToken.MessageChannel <- internal.NewWelcomeMessage(userHasToken)

	tmpUser := *userHasToken
	user := &tmpUser
	user.Token = ""
	// 向所有用户告知新用户到来
	msg := internal.NewUserEnterMessage(user)
	internal.Broadcaster.Broadcast(msg)

	// 将该用户加入广播器的用户列表中
	internal.Broadcaster.UserEntering(user)
	log.Println("user:", nickname, "joins chat")

	// 接受用户消息，从而写入messageChan。这些信息是用来广播给所有用户的
	err = user.ReceiveMessage(req.Context())

	// 用户离开
	internal.Broadcaster.UserLeaving(user) // 将user放入离开队列
	msg = internal.NewUserLeaveMessage(user)
	internal.Broadcaster.Broadcast(msg)
	log.Println("user:", nickname, "leaves chat")

	if err == nil {
		conn.Close(websocket.StatusNormalClosure, "")
	} else {
		log.Println("read from client error:", err)
		conn.Close(websocket.StatusInternalError, "Read from client error")
	}

}
