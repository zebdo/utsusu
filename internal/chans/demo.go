package chans

import (
	"time"
	"github.com/zebdo/utsusu/internal/structs"
)

type DemoSource struct{}

func NewDemoSource() *DemoSource { return &DemoSource{} }
func (d *DemoSource) Name() string { return "demo" }

func (d *DemoSource) FetchThread(board, threadID string) (*structs.Thread, error) {
	return &structs.Thread{
		ID: threadID, Board: board,
		Posts: []structs.Post{
			{ ID: "1", Author: "anon", Content: "Hello, utsusu!", Timestamp: time.Now() },
			{ ID: "2", Author: "anon", Content: "This is a demo thread.", Timestamp: time.Now() },
		},
	}, nil
}

func (d *DemoSource) FetchBoard(board string, limit int) ([]structs.Thread, error) {
	if limit <= 0 { limit = 3 }
	res := make([]structs.Thread, 0, limit)
	for i := 1; i <= limit; i++ {
		res = append(res, structs.Thread{ ID: board + "-" + string(rune('0'+i)), Board: board })
	}
	return res, nil
}
