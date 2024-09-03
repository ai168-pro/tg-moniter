package td

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

var TGClient *tg.Client

const (
	appid = 23406775
	hash  = "5bed7f149b8a2a6d6267c899eeb1e0dc"
)

func InitTGClient() {

	client := telegram.NewClient(appid, hash, telegram.Options{})
	if err := client.Run(context.Background(), func(ctx context.Context) error {
		// It is only valid to use client while this function is not returned
		// and ctx is not cancelled.
		TGClient = client.API()
		codehash, err := SendCode("+8616509107994")
		if err != nil {
			fmt.Println(err, "===11===")
			return err
		}

		var code string
		fmt.Scan(&code)
		SignIn("+8616509107994", code, codehash)

		return nil
	}); err != nil {
		panic(err)
	}
}

func SendCode(phone string) (codehash string, err error) {

	class, err := TGClient.AuthSendCode(context.TODO(), &tg.AuthSendCodeRequest{APIID: appid, APIHash: hash, PhoneNumber: phone})
	if err != nil {
		return
	}

	auth := class.(*tg.AuthSentCode)
	codehash = auth.GetPhoneCodeHash()
	return
}

func SignIn(phone, code, codehash string) (err error) {

	_, err = TGClient.AuthSignIn(context.TODO(), &tg.AuthSignInRequest{PhoneNumber: phone, PhoneCode: code, PhoneCodeHash: codehash})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// auth := authClass.(*tg.AuthAuthorization)
	// fmt.Println(auth)
	fmt.Println(TGClient.AccountGetAccountTTL(context.TODO()))

	// authoriztion, err := TGClient.AuthExportAuthorization(context.Background(), 2)
	// if err != nil {
	// 	fmt.Println(err, "==22==")
	// 	return
	// }
	// fmt.Println(authoriztion.GetBytes())
	// class2, err := TGClient.AuthImportAuthorization(context.TODO(), &tg.AuthImportAuthorizationRequest{ID: 7378749312, Bytes: authoriztion.GetBytes()})
	// if err != nil {
	// 	fmt.Println(err, "===333===")
	// 	return
	// }
	// fmt.Println(class2.String())
	return
}
