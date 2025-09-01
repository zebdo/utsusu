package storage

import "github.com/zebdo/utsusu/internal/structs"

type Storage interface {
	SaveThread(thread structs.Thread) error
	GetThread(id string) (*structs.Thread, error)
	ListThreads(board string) ([]structs.Thread, error)
}
