package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

type Answers struct {
	OriginChannelId string
	FavFood         string
	FavGame         string
	RecordId        int64
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
		{
			Name:  "Record id",
			Value: strconv.FormatInt(a.RecordId, 10),
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

	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sess.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
		if r.Emoji.Name == ("♥️") {
			s.GuildMemberRoleAdd(r.GuildID, r.UserID, "1164907140260057088")
			s.ChannelMessageSend(r.ChannelID, fmt.Sprintf("%v has been added to %v", r.UserID, r.Emoji.Name))
		}
	})

	sess.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
		if r.Emoji.Name == ("♥️") {
			s.GuildMemberRoleRemove(r.GuildID, r.UserID, "1164907140260057088")
			s.ChannelMessageSend(r.ChannelID, fmt.Sprintf("%v has been removed from %v", r.UserID, r.Emoji.Name))
		}
	})

	sess.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		//发送信息的id和session的id一致则不处理信息直接退出
		if m.Author.ID == s.State.User.ID {
			return
		}

		//DM logic
		//如果m.GuildID为空则认为这个消息是direct message
		if m.GuildID == "" {
			UserPromptResponseHandler(db, s, m)
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

		if args[1] == "answers" {
			AnswersHandler(db, s, m)
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

func AnswersHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {
	spl := strings.Split(m.Content, " ")
	if len(spl) < 3 {
		s.ChannelMessageSend(m.ChannelID, "an ID must be provided. Ex: '!gobot answers 1'")
		return
	}
	id, err := strconv.Atoi(spl[2])
	if err != nil {
		log.Fatal(err)
	}

	var recordId int64
	var answerStr string
	var userId int64

	query := "select * from discord_message where id = ?"
	row := db.QueryRow(query, id)

	err = row.Scan(&recordId, &answerStr, &userId)
	if err != nil {
		log.Fatal(err)
	}

	var answers Answers
	err = json.Unmarshal([]byte(answerStr), &answers)
	if err != nil {
		log.Fatal(err)
	}
	answers.RecordId = recordId
	embed := answers.ToMessageEmbed()
	s.ChannelMessageSendEmbed(m.ChannelID, &embed)
}

func UserPromptResponseHandler(db *sql.DB, s *discordgo.Session, m *discordgo.MessageCreate) {
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

		query := "insert into discord_message (payload, user_id) values(?,?)"
		jbytes, err := json.Marshal(answers)
		if err != nil {
			log.Fatal(err)
		}

		res, err := db.Exec(query, string(jbytes), m.ChannelID)
		if err != nil {
			log.Fatal(err)
		}

		lastInsertId, err := res.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}
		answers.RecordId = lastInsertId

		//log.Printf("answers: %v, %v", answers.FavFood, answers.FavGame)
		embed := answers.ToMessageEmbed()
		s.ChannelMessageSendEmbed(answers.OriginChannelId, &embed)

		delete(responses, m.ChannelID)
	}
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
