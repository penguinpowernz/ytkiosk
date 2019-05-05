package ytkiosk

import (
	"bytes"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

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
