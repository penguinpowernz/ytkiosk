package main

import (
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/penguinpowernz/go-ian/util/tell"
	"github.com/penguinpowernz/ytkiosk"
)

var discordChannel = ""
var discordToken = ""

func main() {
	mpv := ytkiosk.NewMPV("/tmp/mpv", "/home/robert/Pictures/100CANON/IMG_4438.JPG")
	go func() {
		for {
			tell.IfErrorf(mpv.Start(), "while starting MPV")
			time.Sleep(time.Second)
		}
	}()

	mpv.WaitForConnection()

	q := ytkiosk.NewQueue()
	eng := ytkiosk.NewEngine(mpv, q)

	api := gin.Default()
	api.Use(static.Serve("/", static.LocalFile("htdocs", false)))

	eng.AttachAPI(api.Group("/api"))
	go api.Run(":8181")

	bot, err := ytkiosk.NewDiscordBot(discordToken, discordChannel)
	if err != nil {
		panic(err)
	}

	bot.Say("I'm awake")
	eng.AttachDiscord(bot)
	defer bot.Kill()

	eng.Run()
}
