package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
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

type Message struct {
	Username, Content string
}

func main() {
	s, _ := discordgo.New(self)
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("( >'-')> Discord selfbot (%s) is ready <('-'< )", r.User.Username)
	})
	if err := s.Open(); err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Telegram bot authorized on account: %s", bot.Self.UserName)

	go health(bot)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-stop:
	default:
		for dm := range dms(s, stop) {
			forward(dm, bot)
		}
	}
	forward(Message{"fwdr", "Shutting down..."}, bot)
	log.Println("Graceful shutdown")
}

func dms(s *discordgo.Session, stop chan os.Signal) <-chan Message {
	out := make(chan Message)
	go func() {
		removeEventHandler := s.AddHandler(handleMessageCreate(out))
		go func() {
			<-stop
			removeEventHandler()
			close(out)
		}()
	}()
	return out
}

func handleMessageCreate(out chan Message) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			log.Fatalf("Cannot open the session: %v", err)
		}
		if c.Type != discordgo.ChannelTypeDM && c.Type != discordgo.ChannelTypeGroupDM {
			return
		} else if m.Author.ID == s.State.User.ID {
			return
		}
		select {
		case out <- Message{m.Author.Username, m.Content}:
		default:
		}
	}
}

func forward(dm Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Message from %s: %s", dm.Username, dm.Content))
	bot.Send(msg)
}

func health(bot *tgbotapi.BotAPI) {
	task := func() {
		forward(Message{"fwdr", "Just letting you know I'm OK!"}, bot)
	}
	gocron.Every(1).Day().At("07:00").Do(task) // 07:00 UTC, i.e. 10:00 MSK
	<-gocron.Start()
}
