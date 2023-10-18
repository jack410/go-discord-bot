package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sess, err := discordgo.New("Bot MTE2NDExOTg1NzExNDEyMDIwMg.GJIDn0.F9lVQRtyj-CPmPlG1h_YYLwG-aR2AX8h9i404Y")
	if err != nil {
		log.Fatal(err)
	}

	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		//发送信息的id和session的id一致则不处理信息直接退出
		if m.Author.ID == s.State.User.ID {
			return
		}

		if m.Content == "hello" {
			_, err = s.ChannelMessageSend(m.ChannelID, "world!")
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = sess.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	fmt.Println("the bot is online!")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
