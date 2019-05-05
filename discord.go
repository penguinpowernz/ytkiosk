package ytkiosk

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/penguinpowernz/go-ian/util/tell"
)

var printEvents = false

func (eng *Engine) AttachDiscord(bot *DiscordBot) {
	_, sub := eng.mpv.Sub()
	go func() {
		tell.Infof("watching events")
		for e := range sub.ch {
			tell.Infof("AND EVENTETETET")
			if printEvents {
				tell.Infof("AND EVENTETETET WAS PRINTED")
				bot.Say("MPV Event: %+v", e)
			}
		}
	}()

	bot.Command("!play", func(args []string) {
		eng.mpv.Play()
		bot.Say("pushed play for you")
	})

	bot.Command("!pause", func(args []string) {
		eng.mpv.Pause()
		bot.Say("pushed pause for you")
	})

	bot.Command("!verbose", func(args []string) {
		if len(args) == 0 {
			return
		}

		if args[0] == "0" {
			printEvents = false
		}

		if args[0] == "1" {
			printEvents = true
		}
	})

	bot.Command("!raw", func(args []string) {
		iargs := []interface{}{}
		for _, a := range args {
			iargs = append(iargs, a)
		}
		res, err := eng.mpv.Call(iargs...)
		bot.Say("raw command returned %+v", res)
		if err != nil {
			bot.Say("it also gave an error: %s", err)
		}
	})

	bot.Command("!get", func(args []string) {
		if len(args) != 1 {
			bot.Say("!get requires exactly one argument")
			return
		}

		res, err := eng.mpv.conn.Get(args[0])
		bot.Say("get returned %+v", res)
		if err != nil {
			bot.Say("it also gave an error: %s", err)
		}
	})

	bot.Command("!set", func(args []string) {
		if len(args) != 2 {
			bot.Say("!get requires exactly one argument")
			return
		}

		err := eng.mpv.conn.Set(args[0], args[1])
		if err != nil {
			bot.Say("failed to set due to an error: %s", err)
			return
		}

		bot.Say("set %s property", args[0])
	})

	bot.Command("!next", func(args []string) {
		eng.mpv.Cancel()
		bot.Say("pushed next for you")
	})

	bot.Command("!now", func(args []string) {
		if len(args) == 0 {
			bot.Say("must specify a URL to play now")
		}

		v := NewVid(args[0])
		title, err := getVideoTitleWithErr(v.URL)
		if err != nil {
			bot.Say("couldn't get the video title for that URL, are you sure it's a YouTube URL?")
			return
		}

		v.Title = title
		eng.q.CutInLine(v)
		eng.mpv.Cancel()
		eng.mpv.Play()
		bot.Say("queuing **%s** immediately", title)
	})

	bot.WithPrefix("http", func(args []string) {
		v := NewVid(args[0])
		title, err := getVideoTitleWithErr(v.URL)
		if err != nil {
			bot.Say("couldn't get the video title for that URL, are you sure it's a YouTube URL?")
			return
		}

		v.Title = title
		eng.q.Add(v)
		bot.Say("queued **%s** for you", title)
	})
}

type DiscordBot struct {
	s       *discordgo.Session
	channel string
}

func NewDiscordBot(token, channel string) (*DiscordBot, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	if err := s.Open(); err != nil {
		return nil, err
	}

	return &DiscordBot{s, channel}, nil
}

func (bot *DiscordBot) Say(text string, args ...interface{}) {
	bot.s.ChannelMessageSend(bot.channel, fmt.Sprintf(text, args...))
}

func (bot *DiscordBot) Command(cmd string, cb func([]string)) {
	bot.s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		if m.Author.ID == s.State.User.ID {
			return
		}

		parts := strings.Split(m.Content, " ")
		_cmd := parts[0]
		args := []string{}

		if _cmd != cmd {
			return
		}

		if len(parts) > 1 {
			args = parts[1:]
		}

		cb(args)
	})
}

func (bot *DiscordBot) WithPrefix(prefix string, cb func([]string)) {
	bot.s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		if m.Author.ID == s.State.User.ID {
			return
		}

		if !strings.HasPrefix(m.Content, prefix) {
			return
		}

		cb(strings.Split(m.Content, " "))
	})
}

func (bot *DiscordBot) Kill() {
	bot.Say("I'm dieyingggggggg!!!!")
	bot.s.Close()
}
