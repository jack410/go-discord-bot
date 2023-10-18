package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const prefix string = "!gobot"

func main() {
	godotenv.Load()

	// 获取 Discord 机器人令牌
	token := os.Getenv("DISCORD_BOT_TOKEN")

	sess, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}

	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		//发送信息的id和session的id一致则不处理信息直接退出
		if m.Author.ID == s.State.User.ID {
			return
		}

		args := strings.Split(m.Content, " ")

		//如果不是以prefix开头则直接不处理
		if args[0] != prefix {
			return
		}

		if args[1] == "hello" {
			_, err = s.ChannelMessageSend(m.ChannelID, "world!")
			if err != nil {
				log.Fatal(err)
			}
		}

		if args[1] == "proverbs" {
			proverbs := []string{
				"Do not communicate by sharing memory; instead, share memory by communicating.",
				"Concurrency is not parallelism.",
				"Channels orchestrate; mutexes serialize.",
				"Don't panic.",
				"Cgo is not Go.",
			}

			selection := rand.Intn(len(proverbs))

			author := discordgo.MessageEmbedAuthor{
				Name: "Rob Pike",
				URL:  "https://go-proverb.github.io",
			}
			embed := discordgo.MessageEmbed{
				Title:  proverbs[selection],
				Author: &author,
			}

			_, err = s.ChannelMessageSendEmbed(m.ChannelID, &embed)
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
