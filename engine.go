package ytkiosk

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/penguinpowernz/go-ian/util/tell"
)

func NewEngine(mpv *MPV, q *Queue) *Engine {
	return &Engine{q: q, mpv: mpv}
}

type Engine struct {
	mu   *sync.Mutex
	mpv  *MPV
	q    *Queue
	curr *Vid
}

func (eng *Engine) Run() {
	for {
		for eng.q.Next() {
			eng.curr = eng.q.Video()
			tell.Debugf("playing URL %s", eng.curr.URL)
			eng.curr.Playing = true
			eng.curr.Played = false
			eng.mpv.Replace(eng.curr.URL)
			time.Sleep(time.Second / 4)

			tell.Debugf("waiting for EOF")
			eng.mpv.WaitForEOF()
			eng.curr.Playing = false
			eng.curr.Played = true

			tell.Debugf("reached EOF")
			time.Sleep(time.Second / 2)
		}

		time.Sleep(time.Second / 2)
	}
}

func (eng *Engine) AttachAPI(api gin.IRouter) {
	api.PUT("/player/pause", func(c *gin.Context) {
		eng.mpv.Pause()
		c.Status(200)
	})

	api.PUT("/player/play", func(c *gin.Context) {
		eng.mpv.Play()
		c.Status(200)
	})

	api.PUT("/player/skip", func(c *gin.Context) {
		eng.mpv.Cancel()
		c.Status(200)
	})

	api.PUT("/player/seek/:pct", func(c *gin.Context) {
		pctS := c.Param("pct")
		pct, err := strconv.Atoi(pctS)
		if err != nil {
			c.AbortWithError(400, err)
			return
		}

		if err := eng.mpv.Seek(pct); err != nil {
			c.AbortWithError(500, err)
			return
		}

		c.Status(200)
	})

	api.POST("/queue", func(c *gin.Context) {
		data := struct {
			U string `json:"url"`
		}{}

		if err := c.BindJSON(&data); err != nil {
			c.AbortWithError(400, err)
			return
		}

		if data.U == "" {
			c.AbortWithError(400, fmt.Errorf("no URL specified"))
			return
		}

		v := NewVid(data.U)
		v.Title = getVideoTitle(data.U)
		eng.q.Add(v)
		c.JSON(200, data)
	})

	api.PUT("/queue", func(c *gin.Context) {
		data := struct {
			U string `json:"url"`
		}{}

		if err := c.BindJSON(&data); err != nil {
			c.AbortWithError(400, err)
			return
		}

		if data.U == "" {
			c.AbortWithError(400, fmt.Errorf("no URL specified"))
			return
		}

		v := NewVid(data.U)
		v.Title = getVideoTitle(data.U)
		eng.q.CutInLine(v)
		eng.mpv.Cancel()
		c.JSON(200, data)
	})

	api.GET("/queue", func(c *gin.Context) {
		if eng.curr != nil {
			eng.curr.Progress = eng.mpv.Progress()
		}
		c.JSON(200, eng.q.All())
	})

	api.DELETE("/queue", func(c *gin.Context) {
		data := struct {
			U string `json:"url"`
		}{}

		if err := c.BindJSON(&data); err != nil {
			c.AbortWithError(400, err)
			return
		}

		if data.U == "" {
			c.AbortWithError(400, fmt.Errorf("no URL specified"))
			return
		}

		if eng.curr.URL == data.U {
			eng.mpv.Cancel()
		}

		v := &Vid{URL: data.U}
		cnt := eng.q.Remove(v)
		c.JSON(200, map[string]int{"deleted": cnt})
	})
}

func getVideoTitle(url string) string {
	http.DefaultClient.Timeout = 30 * time.Second
	res, err := http.Get(url)
	if err != nil {
		return err.Error()
	}

	b := bytes.NewBuffer([]byte{})
	b.ReadFrom(res.Body)

	r := regexp.MustCompile("<title>(.*)</title>")
	x := r.FindStringSubmatch(b.String())
	return strings.Replace(x[1], " - YouTube", "", -1)
}
