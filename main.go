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

type Answers struct {
	OriginChannelId string
	FavFood         string
	FavGame         string
}

func (a *Answers) ToMessageEmbed() discordgo.MessageEmbed {
	fields := []*discordgo.MessageEmbedField{
		{
			Name:  "Favorite food",
			Value: a.FavFood,
		},
		{
			Name:  "Favorite game",
			Value: a.FavFood,
		},
	}

	return discordgo.MessageEmbed{
		Title:  "New responses!",
		Fields: fields,
	}
}

var responses map[string]Answers = map[string]Answers{}

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

		//DM logic
		//如果m.GuildID为空则认为这个消息是direct message
		if m.GuildID == "" {
			answers, ok := responses[m.ChannelID]
			if !ok {
				return
			}
			if answers.FavFood == "" {
				answers.FavFood = m.Content

				s.ChannelMessageSend(m.ChannelID, "Great! What's your favorite game now?")

				responses[m.ChannelID] = answers
				return
			} else {
				answers.FavGame = m.Content
				//log.Printf("answers: %v, %v", answers.FavFood, answers.FavGame)
				embed := answers.ToMessageEmbed()
				sess.ChannelMessageSendEmbed(answers.OriginChannelId, &embed)

				delete(responses, m.ChannelID)
			}
		}

		//server logic
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

		if args[1] == "prompt" {
			UserPromptHandler(s, m)
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

func UserPromptHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	//user channel
	channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		log.Panic(err)
	}

	//if the user is already answers questions, ignore it, otherwise ask questions.
	if _, ok := responses[channel.ID]; !ok {
		responses[channel.ID] = Answers{
			OriginChannelId: m.ChannelID,
			FavFood:         "",
			FavGame:         "",
		}
		s.ChannelMessageSend(channel.ID, "Hey there! Here are some questions")
		s.ChannelMessageSend(channel.ID, "What's your favorite food?")
	} else {
		s.ChannelMessageSend(channel.ID, "We're still waiting...")
	}
}
