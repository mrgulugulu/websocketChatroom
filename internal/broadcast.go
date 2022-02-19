package internal

import (
	"expvar"
	"fmt"
	"log"
	"wsChatroom/global"
)

func init() {
	expvar.Publish("message_queue", expvar.Func(calcMessageQueueLen))
}
func calcMessageQueueLen() interface{} {
	fmt.Println("===len=:", len(Broadcaster.messageChannel))
	return len(Broadcaster.messageChannel)
}

type broadcaster struct {
	users map[string]*User

	// 所有channel统一 管理，不导出
	enteringChannel chan *User
	leavingChannel  chan *User
	messageChannel  chan *Message

	//
	checkUserChannel      chan string
	checkUserCanInChannel chan bool

	requestUsersChannel chan struct{}
	usersChannel        chan []*User
}

func (b *broadcaster) Start() {
	for {
		select {
		// 有用户进入
		case user := <-b.enteringChannel:
			b.users[user.NickName] = user // 注册
			OfflineProcessor.Send(user)

		case user := <-b.leavingChannel:
			// 在注册列表中删除
			delete(b.users, user.NickName)
			// 关闭对应user的chan,避免goroutine泄露
			user.CloseMessageChannel()

		case msg := <-b.messageChannel:
			for _, user := range b.users {
				// 判断是否给自己，是的话就跳过
				if user.UID == msg.User.UID {
					continue
				}
				// 将消息发给每一个人的msgchan中
				user.MessageChannel <- msg
			}
			// 对信息进行保存
			OfflineProcessor.Save(msg)
		case nickname := <-b.checkUserChannel:
			// 判断是否已经注册了
			if _, ok := b.users[nickname]; ok {
				b.checkUserCanInChannel <- false
			} else {
				b.checkUserCanInChannel <- true
			}
		case <-b.requestUsersChannel:
			userList := make([]*User, 0, len(b.users))
			for _, user := range b.users {
				userList = append(userList, user)
			}

			b.usersChannel <- userList
		}
	}
}

// 饿汉式单例模式，这里作为唯一单例
var Broadcaster = &broadcaster{
	users: make(map[string]*User),

	enteringChannel: make(chan *User),
	leavingChannel:  make(chan *User),
	messageChannel:  make(chan *Message, global.MessageQueueLen),

	checkUserChannel:      make(chan string),
	checkUserCanInChannel: make(chan bool),

	requestUsersChannel: make(chan struct{}),
	usersChannel:        make(chan []*User),
}

// 用chan作为trigger，建议用这种专用通道来进行触发
func (b *broadcaster) UserEntering(u *User) {
	b.enteringChannel <- u
}
func (b *broadcaster) UserLeaving(u *User) {
	b.leavingChannel <- u
}

// 这个chan是属于广播器的，所以是向所有用户发送
func (b *broadcaster) Broadcast(msg *Message) {
	if len(b.messageChannel) >= global.MessageQueueLen {
		log.Println("broadcast queue 满了")
	}
	b.messageChannel <- msg
}

func (b *broadcaster) CanEnterRoom(nickname string) bool {
	b.checkUserChannel <- nickname
	return <-b.checkUserCanInChannel
}

func (b *broadcaster) GetUserList() []*User {
	b.requestUsersChannel <- struct{}{}
	return <-b.usersChannel
}
