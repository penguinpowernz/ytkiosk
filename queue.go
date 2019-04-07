package ytkiosk

import (
	"container/list"
	"encoding/json"
	"io/ioutil"
)

func NewQueue() *Queue {
	l := list.New()
	l.Init()
	return &Queue{l, nil}
}

type Queue struct {
	l    *list.List
	curr *list.Element
}

func (q *Queue) CutInLine(v *Vid) {
	if q.curr == nil {
		q.Add(v)
		return
	}

	q.l.InsertAfter(v, q.curr)
}

func (q *Queue) Next() bool {
	if q.curr == nil {
		e := q.l.Front()
		if e == nil {
			return false
		}

		q.curr = e
		return true
	}

	n := q.curr.Next()
	if n == nil {
		return false
	}

	q.curr = n
	return true
}

func (q *Queue) Video() *Vid {
	return q.curr.Value.(*Vid)
}

func (q *Queue) Add(v *Vid) {
	q.l.PushBack(v)
}

func (q *Queue) Clean() {
	for {
		e := q.l.Back()
		if e == nil {
			break
		}

		v := e.Value.(*Vid)
		if v.Played {
			q.l.Remove(e)
		}
	}
}

func (q *Queue) All() []*Vid {
	vids := []*Vid{}
	e := q.l.Front()
	if e == nil {
		return vids
	}

	vids = append(vids, e.Value.(*Vid))

	for {
		e = e.Next()
		if e == nil {
			break
		}
		vids = append(vids, e.Value.(*Vid))
	}

	return vids
}

func (q *Queue) Remove(v *Vid) int {
	var cnt int
	e := q.l.Front()
	if e == nil {
		return cnt
	}

	for {
		e = e.Next()
		if e == nil {
			break
		}

		if e.Value.(*Vid).URL == v.URL {
			e = e.Next()
			q.l.Remove(e.Prev())
			cnt++
		}
	}

	return cnt
}

func (q *Queue) Prev() bool {
	n := q.curr.Prev()
	if n == nil {
		return false
	}

	q.curr = n
	return true
}

func (q *Queue) Save(fn string) error {
	data, err := json.Marshal(q.All())
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fn, data, 0644)
}

func (q *Queue) Load(fn string) error {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}

	vids := []*Vid{}
	if err := json.Unmarshal(data, &vids); err != nil {
		return err
	}

	for _, v := range vids {
		q.Add(v)
	}

	return nil
}

func NewVid(url string) *Vid {
	return &Vid{URL: url}
}

type Vid struct {
	URL      string `json:"url"`
	Title    string `json:"title"`
	Played   bool   `json:"played"`
	Playing  bool   `json:"playing"`
	Progress int    `json:"progress"`
}
