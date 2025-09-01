package storage

import "github.com/zebdo/utsusu/internal/core"

type Storage interface {
	SaveThread(thread core.Thread) error
	GetThread(id string) (*core.Thread, error)
	ListThreads(board string) ([]core.Thread, error)
}
