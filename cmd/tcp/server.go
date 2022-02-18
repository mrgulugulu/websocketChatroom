package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

type User struct {
	ID             int
	Addr           string
	EnterAt        time.Time   // 进入时间
	MessageChannel chan string // 当前发送信息的通道
}

func (u User) String() string {
	return fmt.Sprintf("%v, %v, %v", u.Addr, u.ID, u.EnterAt)
}

type Message struct {
	OwnerID int
	Content string
}

var (
	userID int
	// 记录新用户到来
	enteringChannel = make(chan *User)
	// 记录用户离开
	leavingChannel = make(chan *User)
	// 广播专用的用户普通消息
	messageChannel = make(chan Message, 8)
)

func main() {
	listener, err := net.Listen("tcp", ":2020")
	if err != nil {
		panic(err)
	}
	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConn(conn)
	}
}
func handleConn(conn net.Conn) {
	defer conn.Close()

	user := &User{
		ID:             GenUserID(),
		Addr:           conn.RemoteAddr().String(),
		EnterAt:        time.Now(),
		MessageChannel: make(chan string, 8),
	}
	// 开一个goroutine用于写操作，读写之间通过channel来通信
	go sendMessage(conn, user.MessageChannel)

	// 给当前用户发送欢迎信息
	user.MessageChannel <- "Welcome, " + user.String()
	msg := Message{
		OwnerID: user.ID,
		Content: "user:`" + strconv.Itoa(user.ID) + "` has enter",
	}
	messageChannel <- msg

	// 记录到全局用户列表中，没有用锁
	enteringChannel <- user

	var userActive = make(chan struct{})

	// 控制超时
	go func() {
		d := 1 * time.Minute
		timer := time.NewTicker(d)
		for {
			select {
			case <-timer.C:
				conn.Close() //超时就直接中断
			case <-userActive:
				timer.Reset(d)
			}
		}
	}()
	// 循环读取用户输入
	input := bufio.NewScanner(conn)
	for input.Scan() {
		msg.Content = strconv.Itoa(user.ID) + ":" + input.Text()
		messageChannel <- msg
	}
	if err := input.Err(); err != nil {
		log.Println("读取错误：", err)
	}

	// 用户活跃
	userActive <- struct{}{}

	// 用户离开
	leavingChannel <- user
	msg.Content = "user:`" + strconv.Itoa(user.ID) + "` has left"
	messageChannel <- msg

}

// 有可能存在goroutine泄露的问题
func sendMessage(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg) // 从chan读取后写入conn中
	}
}

// 记录聊天室用户，并进行信息广播
func broadcaster() {
	users := make(map[*User]struct{})

	for {
		select {
		// 新用户进入
		case user := <-enteringChannel:
			users[user] = struct{}{}
		// 用户离开
		case user := <-leavingChannel:
			delete(users, user)
			// 注销用户时，除从map中删除用户外，还要关闭user的chan，避免goroutine泄露
			close(user.MessageChannel)
		case msg := <-messageChannel:
			// 给所有在线用户发送消息
			for user := range users {
				if user.ID == msg.OwnerID {
					continue
				}
				user.MessageChannel <- msg.Content
			}
		}
	}
}

var (
	globalID int
	idLocker sync.Mutex
)

func GenUserID() int {
	idLocker.Lock()
	defer idLocker.Unlock()
	globalID++
	return globalID
}
