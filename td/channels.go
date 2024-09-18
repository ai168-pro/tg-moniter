package td

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"tg-moniter/sensitive"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type ChannelMessageService struct {
	channels         []*tg.Channel
	channelsMaxMsgId map[int64]int
	ctx              context.Context
	userIDSet        map[int64]*tg.User
	locker           sync.Mutex
	msg              map[int64][]*tg.Message

	client     *telegram.Client
	groupNum   int
	msgListNum int
	validate   int
}

func NewChannelMessageService(glimit, msglimit, validate int, ctx context.Context, client *telegram.Client) *ChannelMessageService {
	return &ChannelMessageService{userIDSet: make(map[int64]*tg.User), msg: make(map[int64][]*tg.Message), validate: validate, groupNum: glimit, msgListNum: msglimit, ctx: ctx, channelsMaxMsgId: make(map[int64]int), client: client}
}

func (cm *ChannelMessageService) ready() {
	var err error
	err = cm.getChannelsDialogs()
	if err != nil {
		fmt.Println("初始化失败")
		log.Fatalln(err.Error())
	}
}

func (cm *ChannelMessageService) getChannelsDialogs() (err error) {
	messageDialogs, err := cm.client.API().MessagesGetDialogs(cm.ctx, &tg.MessagesGetDialogsRequest{
		OffsetDate: 0,
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      cm.groupNum, // 获取对话所需的个数
	})
	if err != nil {
		return
	}

	dialogs := messageDialogs.(*tg.MessagesDialogsSlice)
	for _, chat := range dialogs.Chats {
		switch ch := chat.(type) {
		case *tg.Channel: // 判断是否为群组或超级群组
			if ch.Megagroup { // Megagroup 表示超级群组
				log.Printf("Group ID: %d, Title: %s", ch.ID, ch.Title)
				cm.channels = append(cm.channels, ch)
				// // 获取群组的用户名（如果存在）
				// if ch.Username != "" {
				// 	log.Printf("Group Username: %s", ch.Username)
				// } else {
				// 	log.Printf("Group has no username")
				// }
			}
		default:
			// 忽略非群组对话
		}
	}
	return
}

func (cm *ChannelMessageService) getMessage() (err error) {
	rand.Seed(time.Now().UnixNano())
	cm.ctx.Done()

	for {
		for _, channel := range cm.channels {
			time.Sleep(time.Millisecond * 100)
			select {
			case <-cm.ctx.Done():
				log.Fatalln("强制退出")
			default:
			}
			msgsClass, err := cm.client.API().MessagesGetHistory(cm.ctx, &tg.MessagesGetHistoryRequest{Peer: &tg.InputPeerChannel{ChannelID: channel.ID, AccessHash: channel.AccessHash}, OffsetID: 0, Limit: cm.msgListNum})
			if err != nil {
				fmt.Printf("查询最近群组消息失败；channid: %d ; 群组名称： %s; accesshash: %d;err: %s \n", channel.ID, channel.Title, channel.AccessHash, err.Error())
				continue
			}
			msgs := msgsClass.(*tg.MessagesChannelMessages)

			var firstCount int
			for _, v := range msgs.Messages {
				msg, ok := v.(*tg.Message)
				if !ok {
					continue
				}
				if msg.ID <= int(cm.channelsMaxMsgId[channel.ID]) {
					// fmt.Printf("该群组无最新消息；channid: %d ; 群组名称： %s;上次查询ID %d ;本次查询最大ID：%d; \n", channel.ID, channel.Title, cm.channelsMaxMsgId[channel.ID], msg.ID)
					break
				}
				firstCount++
				// 记录此次查询到的群组最大消息记录ID
				if firstCount == 1 {
					cm.channelsMaxMsgId[channel.ID] = msg.ID
				}
				if msg.Message == "" {
					continue
				}
				unsensitive, word := sensitive.SensitiveWord.Validate(msg.Message)
				if !unsensitive || cm.validate > 0 {
					var user = &tg.User{}
					if msg.FromID == nil {
						continue
					}
					fromId, ok := msg.FromID.(*tg.PeerUser)
					if !ok {
						fmt.Println("跳过不是用户的信息")
						continue
					}
					for _, userClass := range msgs.Users {
						usr := userClass.(*tg.User)
						if usr.ID == fromId.UserID {
							user = usr
							break
						}
					}
					if user.Bot {
						if len(msg.Message) > 27 {
							fmt.Println("忽略掉机器人消息 : --", msg.Message[:27])
						} else {
							fmt.Println("忽略掉机器人消息", msg.Message)
						}
						continue
					}
					if _, ok := cm.userIDSet[user.ID]; !ok {
						if user.Username == "" {
							users, err := client.API().UsersGetUsers(context.TODO(), []tg.InputUserClass{&tg.InputUser{UserID: user.ID, AccessHash: user.AccessHash}})
							if err == nil {
								if len(users) > 0 {
									user1 := users[0].(*tg.User)
									user.Username = user1.Username
								}
							}

						}
						cm.locker.Lock()
						cm.userIDSet[user.ID] = user
						cm.locker.Unlock()
						fmt.Printf("加入记录 ； 用户名：%s ；昵称：%s;群组：%s;内容： %s;", user.Username, user.FirstName+user.LastName, channel.Title, msg.Message)
						writeSensitiveWordToFile(msg, user, channel, word)
					}
				}
			}
		}

		time.Sleep(time.Second * 5)
		waittimes := rand.Intn(10)
		for i := 0; i <= waittimes; i++ {
			select {
			case <-cm.ctx.Done():
				log.Fatalln("强制退出")
			default:
			}
		}
	}

}
