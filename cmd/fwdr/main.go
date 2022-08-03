package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	self     string = os.Getenv("SELF")
	botToken string = os.Getenv("BOT_TOKEN")
	chatID   int64
)

func init() {
	i, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	if err != nil {
		panic(err)
	}
	chatID = i
}

func main() {
	s, _ := discordgo.New(self)
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Discord selfbot (%s) is ready", r.User.Username)
	})

	bot := bot(botToken)
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			log.Fatalf("Cannot open the session: %v", err)
		}
		if c.Type != discordgo.ChannelTypeDM && c.Type != discordgo.ChannelTypeGroupDM {
			return
		} else if m.Author.ID == s.State.User.ID {
			return
		}
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("New Discord message from %s: %s", m.Author, m.Content))
		bot.Send(msg)
	})

	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			s.UpdateStatusComplex(discordgo.UpdateStatusData{Status: "online"})
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}

func bot(token string) *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Telegram bot authorized on account %s", bot.Self.UserName)
	return bot
}
