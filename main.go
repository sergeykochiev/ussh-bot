package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/sergeykochiev/ussh/telegram"

	"github.com/joho/godotenv"
)

type Shell struct {
	cmd     *exec.Cmd
	in      io.WriteCloser
	out     io.ReadCloser
	err     io.ReadCloser
	outChan chan string
}

func (sh *Shell) Init() (err error) {
	sh.cmd = exec.Command("/usr/bin/sh")
	sh.outChan = make(chan string)
	sh.out, err = sh.cmd.StdoutPipe()
	if err != nil {
		return
	}
	sh.err, err = sh.cmd.StderrPipe()
	if err != nil {
		return
	}
	sh.in, err = sh.cmd.StdinPipe()
	return
}

func (sh *Shell) Start() {
	go sh.cmd.Run()
	go sh.StartReading(sh.out)
	go sh.StartReading(sh.err)
}

func (sh *Shell) Run(cmd string) error {
	_, err := sh.in.Write([]byte(cmd))
	return err
}

func (sh *Shell) StartReading(r io.ReadCloser) {
	reader := bufio.NewReader(r)
	for {
		output, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Error reading output: ", err)
		}
		sh.outChan <- output[:len(output)-1]
	}
}

func (sh *Shell) Output() string {
	return <-sh.outChan
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load .env")
	}

	bot := telegram.Bot{}
	bot.AllowedUpdates = []string{"message"}

	bot.Token = os.Getenv("BOT_TOKEN")
	if len(bot.Token) == 0 {
		log.Fatal("Token is not provided in .env")
	}

	targetChatId, err := strconv.Atoi(os.Getenv("TARGET_CHAT_ID"))
	if err != nil {
		log.Fatal("Failed to load user id from .env: invalid value")
	}

	shell := Shell{}

	err = shell.Init()
	if err != nil {
		log.Fatal("Failed to start shell: ", err)
	}

	shell.Start()

	go func() {
		for {
			bot.SendMessage(targetChatId, shell.Output())
		}
	}()

	for {
		updates, err := bot.GetUpdates()
		if err != nil {
			log.Print(err)
			continue
		}
		log.Print("Received updates")

		updatesArr := updates.Result
		if len(updatesArr) == 0 {
			continue
		}

		for _, update := range updatesArr {
			bot.LastUpdateId = update.UpdateId + 1

			message := update.Message

			chatId := message.Chat.Id
			if chatId != targetChatId {
				continue
			}

			text := message.Text
			if text[0] != '$' {
				continue
			}
			command := text[2:]

			log.Print("Running command: ", command)

			err = shell.Run(command + "\n")

			if err != nil {
				log.Print("Error running command: ", err)
			}
		}
	}
}
