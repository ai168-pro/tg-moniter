package td

import (
	"context"
	"fmt"
	"os"
	"tg-moniter/sensitive"

	"github.com/gotd/td/examples"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	updhook "github.com/gotd/td/telegram/updates/hook"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// func init() {
// 	wd, _ := os.Getwd()
// 	_, err := os.Stat(wd + "/" + "result.txt")
// 	if os.IsNotExist(err) {
// 		result, err = os.Create(wd + "/" + "result.txt")
// 	} else {
// 		result, err = os.OpenFile(wd+"/"+"result.txt", os.O_RDWR|os.O_APPEND, os.ModeAppend)
// 		if err != nil {
// 			return
// 		}
// 	}
// }

func Run(ctx context.Context, phone string) error {
	log, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.InfoLevel), zap.AddStacktrace(zapcore.FatalLevel))
	defer func() { _ = log.Sync() }()

	d := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: d,
		Logger:  log.Named("gaps"),
	})

	flow := auth.NewFlow(examples.Terminal{PhoneNumber: phone}, auth.SendCodeOptions{})
	wd, _ := os.Getwd()
	// /Users/nguyenoanh/Desktop/session/
	client := telegram.NewClient(appid, hash, telegram.Options{SessionStorage: &session.FileStorage{Path: wd + "/" + phone + ".json"}, Logger: log,
		UpdateHandler: gaps,
		Middlewares: []telegram.Middleware{
			updhook.UpdateHook(gaps.Handle),
		}})

	// Setup message update handlers.
	d.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
		// log.Info("Channel message", zap.Any("message", update.Message))

		if msg, ok := update.Message.(*tg.Message); ok {
			log.Info("收到消息：", zap.String("message", msg.Message))
			unsensitive, word := sensitive.SensitiveWord.Validate(msg.Message)
			if !unsensitive {
				log.Info("命中关键词 ：", zap.Any("message", msg.Message), zap.String("关键词", word))

				var user = &tg.User{}
				var channel = &tg.Channel{}
				// peerUser := msg.FromID.(*tg.PeerUser)
				// user = e.Users[peerUser.UserID]
				if peerUser, ok := msg.FromID.(*tg.PeerUser); ok {
					user = e.Users[peerUser.UserID]
				}
				if peerChannel, ok := msg.PeerID.(*tg.PeerChannel); ok {
					channel = e.Channels[peerChannel.ChannelID]
				}

				if !user.Bot {
					writeSensitiveWordToFile(msg, user, channel, word)
				}
			}
		}

		// message := update.Message.(*tg.Message)
		// fmt.Println(e.Users, "===111===")
		// fmt.Println(e.Channels, "===222===")
		return nil
	})
	d.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
		log.Info("Message", zap.Any("message", update.Message))
		return nil
	})

	return client.Run(ctx, func(ctx context.Context) error {
		// Perform auth if no session is available.
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return errors.Wrap(err, "auth")
		}

		// Fetch user info.
		user, err := client.Self(ctx)
		if err != nil {
			return errors.Wrap(err, "call self")
		}

		return gaps.Run(ctx, client.API(), user.ID, updates.AuthOptions{
			OnStart: func(ctx context.Context) {
				log.Info("Gaps started")
			},
		})
	})
}

func writeSensitiveWordToFile(msg *tg.Message, user *tg.User, channel *tg.Channel, word string) (err error) {

	wd, _ := os.Getwd()
	result, err := os.OpenFile(wd+"/"+"result.txt", os.O_RDWR|os.O_APPEND, os.ModeAppend)
	if err != nil {
		fmt.Println(err, "=====openfile====")
		return
	}
	defer result.Close()
	result.WriteString(fmt.Sprintf("消息ID：%d ；群组名称： %s ; 发送用户者： %s ;手机号码： %s ; 关键词：%s ；全文: %s ； \n", msg.ID, channel.Title, user.FirstName+user.LastName, user.Phone, word, msg.Message))
	result.WriteString("============================\n")
	return
}
