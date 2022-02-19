package internal

import (
	"container/ring"
	"github.com/spf13/viper"
)

type offlineProcessor struct {
	n int

	// 保存所有用户最近的n条消息
	recentRing *ring.Ring

	// 保存某个用户的离线消息n条
	userRing map[string]*ring.Ring
}

var OfflineProcessor = newOfflineProcessor()

func newOfflineProcessor() *offlineProcessor {
	n := viper.GetInt("offline-num")

	return &offlineProcessor{
		n:          n,
		recentRing: ring.New(n),
		userRing:   make(map[string]*ring.Ring),
	}
}

func (o *offlineProcessor) Save(msg *Message) {
	if msg.Type != MsgTypeNormal {
		return
	}
	o.recentRing.Value = msg
	o.recentRing = o.recentRing.Next()

	for _, nickname := range msg.Ats {
		nickname = nickname[1:]
		var (
			r  *ring.Ring
			ok bool
		)
		// 如果不存在该用户，就给它创建一个ring
		if r, ok = o.userRing[nickname]; !ok {
			r = ring.New(o.n)
		}
		r.Value = msg
		// 取后一个的nickname的ring出来
		o.userRing[nickname] = r.Next()
	}
}

func (o *offlineProcessor) Send(user *User) {
	o.recentRing.Do(func(value interface{}) {
		if value != nil {
			user.MessageChannel <- value.(*Message) // 整个msg取出来
		}
	})
	if user.isNew {
		return
	}
	// 取个人的消息
	if r, ok := o.userRing[user.NickName]; ok {
		r.Do(func(value interface{}) {
			if value != nil {
				user.MessageChannel <- value.(*Message)
			}
		})
		delete(o.userRing, user.NickName)
	}
}
