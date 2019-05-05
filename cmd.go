package ytkiosk

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/DexterLB/mpvipc"
	"github.com/penguinpowernz/go-ian/util/tell"
)

func NewMPV(socket, image string) *MPV {
	mpv := &MPV{
		socket:   socket,
		backdrop: image,
	}

	return mpv
}

type MPV struct {
	socket   string
	backdrop string
	*exec.Cmd
	killChan          chan struct{}
	conn              *mpvipc.Connection
	events            []mpvipc.Event
	Subscribers       []Subscription
	listenCancel      chan struct{}
	playbackCancelled bool
	connected         bool
}

func (mpv *MPV) ResetCommand() {
	mpv.Cmd = exec.Command("/usr/bin/mpv", "--fs", "--fs-screen=0", "--input-unix-socket="+mpv.socket, mpv.backdrop, "--keep-open=yes", "--keep-open-pause=no", "--force-window=yes", "--image-display-duration=inf")
	mpv.Stdout = os.Stdout
	mpv.Stderr = os.Stderr
}

func (mpv *MPV) Sub() (int, Subscription) {
	s := Subscription{make(chan mpvipc.Event)}
	mpv.Subscribers = append(mpv.Subscribers, s)
	return len(mpv.Subscribers), s
}

func (mpv *MPV) Unsub(id int) {
	tell.Infof("unsubbing")
	close(mpv.Subscribers[id-1].ch)
	tell.Infof("wiating")
}

func (mpv *MPV) Publish(ev mpvipc.Event) {
	for _, s := range mpv.Subscribers {
		if s.ch == nil {
			continue
		}

		s.Publish(ev)
	}
}

func (mpv *MPV) Start() error {
	mpv.killChan = make(chan struct{})
	mpv.ResetCommand()

	var rerr error
	go func() {
		if err := mpv.Cmd.Run(); err != nil {
			tell.IfErrorf(err, "the player cmd stopped")
			rerr = err
			mpv.Kill()
		}
	}()

	defer func() {
		if mpv.Process == nil {
			return
		}
		mpv.Process.Kill()
	}()

	time.Sleep(time.Second / 2)

	mpv.conn = mpvipc.NewConnection(mpv.socket)
	if err := mpv.conn.Open(); err != nil {
		return err
	}

	go mpv.Listen()

	mpv.connected = true
	<-mpv.killChan
	mpv.connected = false
	close(mpv.listenCancel)
	mpv.conn.Close()
	tell.Debugf("%T %+v", rerr, rerr)
	return rerr
}

func (mpv *MPV) Kill() {
	close(mpv.killChan)
}

func (mpv *MPV) Listen() {
	var events chan *mpvipc.Event
	events, mpv.listenCancel = mpv.conn.NewEventListener()

	for event := range events {
		tell.Debugf("EVENT: %+v", event)
		mpv.events = append(mpv.events, *event)

		for _, sub := range mpv.Subscribers {
			sub.Publish(*event)
		}
	}
}

func (mpv *MPV) WaitForConnection() {
	for !mpv.connected {
	}
}

func (mpv *MPV) WaitForEOF() {
	mpv.playbackCancelled = false

	for {
		if mpv.playbackCancelled {
			return
		}

		v, err := mpv.conn.Get("eof-reached")
		if err != nil && strings.Contains(err.Error(), "unavailable") {
			time.Sleep(time.Second / 2)
			continue
		}

		if err != nil {
			tell.IfErrorf(err, "failed to get property eof-reached")
			break
		}

		if v.(bool) {
			break
		}

		time.Sleep(time.Second / 3)
	}
}

func (mpv *MPV) Cancel() {
	mpv.playbackCancelled = true
}

func (mpv *MPV) Seek(pct int) error {
	return mpv.conn.Set("percent-pos", float64(pct))
}

func (mpv *MPV) Progress() int {
	v, err := mpv.conn.Get("percent-pos")
	if err != nil {
		return 0
	}
	return int(v.(float64))
}

func (mpv *MPV) Pause() {
	a := mpv.conn.Set("pause", true)
	tell.Debugf("%+v", a)
}

func (mpv *MPV) Play() {
	a := mpv.conn.Set("pause", false)
	tell.Debugf("%+v", a)
}

func (mpv *MPV) Replace(url string) {
	mpv.conn.Call("loadfile", url, "replace")
}

func (mpv *MPV) Queue(url string) {
	mpv.conn.Call("loadfile", url, "append-play")
}

func (mpv *MPV) Call(args ...interface{}) (interface{}, error) {
	return mpv.conn.Call(args...)
}

func (mpv *MPV) Events() []mpvipc.Event {
	return mpv.events
}

type Subscription struct {
	ch chan mpvipc.Event
}

func (s *Subscription) Publish(ev mpvipc.Event) error {
	if _, ok := <-s.ch; !ok {
		return errors.New("Topic has been closed")
	}

	s.ch <- ev

	return nil
}
