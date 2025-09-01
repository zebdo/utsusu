package chans

import "github.com/zebdo/utsusu/internal/core"

type ChanSource interface {
	FetchThread(board, threadID string) (*core.Thread, error)
	FetchBoard(board string, limit int) ([]core.Thread, error)
	Name() string
}