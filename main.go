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
	if err := td.Run(ctx, phone); err != nil {
		panic(err)
	}

}
