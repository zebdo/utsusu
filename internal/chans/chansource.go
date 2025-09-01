package chans

import "github.com/zebdo/utsusu/internal/structs"

type ChanSource interface {
	FetchThread(board, threadID string) (*structs.Thread, error)
	FetchBoard(board string, limit int) ([]structs.Thread, error)
	Name() string
}