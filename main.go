package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	_ "tg-moniter/sysinit"
	"tg-moniter/td"

	"gopkg.in/ini.v1"
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cur, _ := os.Getwd()
	cfg, err := ini.Load(cur + "/my.ini")
	if err != nil {
		log.Fatalf("Fail to read file: %v", err)
	}

	phone := cfg.Section("").Key("phone").String()
	groupNum := cfg.Section("").Key("groupnum").MustInt(10)
	msgNum := cfg.Section("").Key("msgnum").MustInt(50)
	validate := cfg.Section("").Key("allmsg").MustInt(0)
	if err := td.Run2(ctx, phone, groupNum, msgNum, validate); err != nil {
		panic(err)
	}

}
