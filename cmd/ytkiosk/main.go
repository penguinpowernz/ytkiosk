package main

import (
	"flag"
	"io/ioutil"
	"time"

	"github.com/ghodss/yaml"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/penguinpowernz/go-ian/util/tell"
	"github.com/penguinpowernz/ytkiosk"
)

type config struct {
	DiscordChannel string `json:"discord_channel"`
	DiscordToken   string `json:"discord_token"`
	BackdropImage  string `json:"backdrop_image"`
}

var cfgFile string

func main() {
	flag.StringVar(&cfgFile, "-c", "config.yml", "the config file to use")
	flag.Parse()

	data, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		tell.IfFatalf(err, "while reading config")
	}

	var cfg config
	err = yaml.Unmarshal(data, &cfg)
	tell.IfFatalf(err, "while parsing YAML")

	mpv := ytkiosk.NewMPV("/tmp/mpv", cfg.BackdropImage)
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

	if cfg.DiscordToken != "" {
		bot, err := ytkiosk.NewDiscordBot(cfg.DiscordToken, cfg.DiscordChannel)
		if err != nil {
			panic(err)
		}

		bot.Say("I'm awake")
		eng.AttachDiscord(bot)
		defer bot.Kill()
	}

	eng.Run()
}
