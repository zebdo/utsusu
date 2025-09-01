package chans

import (
	"time"
	"github.com/zebdo/utsusu/internal/core"
)

type DemoSource struct{}

func NewDemoSource() *DemoSource { return &DemoSource{} }
func (d *DemoSource) Name() string { return "demo" }

func (d *DemoSource) FetchThread(board, threadID string) (*core.Thread, error) {
	return &core.Thread{
		ID: threadID, Board: board,
		Posts: []core.Post{
			{ ID: "1", Author: "anon", Content: "Hello, utsusu!", Timestamp: time.Now() },
			{ ID: "2", Author: "anon", Content: "This is a demo thread.", Timestamp: time.Now() },
		},
	}, nil
}

func (d *DemoSource) FetchBoard(board string, limit int) ([]core.Thread, error) {
	if limit <= 0 { limit = 3 }
	res := make([]core.Thread, 0, limit)
	for i := 1; i <= limit; i++ {
		res = append(res, core.Thread{ ID: board + "-" + string(rune('0'+i)), Board: board })
	}
	return res, nil
}
