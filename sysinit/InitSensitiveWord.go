package sysinit

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"tg-moniter/sensitive"
)

func InitSensitiveWord() {

	cur, _ := os.Getwd()
	wordByte, err := ioutil.ReadFile(cur + "/word.txt")
	if err != nil {
		log.Fatalln(err)
	}
	for _, v := range strings.Split(string(wordByte), ",") {
		sensitive.SensitiveWord.AddWord(v)
	}
}
