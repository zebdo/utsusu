package core

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/zebdo/utsusu/internal/chans"
	"github.com/zebdo/utsusu/internal/storage"
)

type Watch struct {
	Source   string
	Board    string
	ThreadID string
	Every    time.Duration
}

type Archiver struct {
	store   storage.Storage
	sources map[string]chans.ChanSource

	mu      sync.RWMutex
	watches map[string]Watch // key: source|board|thread
}

func NewArchiver(store storage.Storage, sources map[string]chans.ChanSource) *Archiver {
	return &Archiver{
		store: store,
		sources: sources,
		watches: make(map[string]Watch),
	}
}

func (a *Archiver) key(w Watch) string { return w.Source + "|" + w.Board + "|" + w.ThreadID }

func (a *Archiver) AddWatch(w Watch) {
	a.mu.Lock(); defer a.mu.Unlock()
	a.watches[a.key(w)] = w
}

func (a *Archiver) RemoveWatch(w Watch) {
	a.mu.Lock(); defer a.mu.Unlock()
	delete(a.watches, a.key(w))
}

func (a *Archiver) ListWatches() []Watch {
	a.mu.RLock(); defer a.mu.RUnlock()
	res := make([]Watch, 0, len(a.watches))
	for _, w := range a.watches { res = append(res, w) }
	return res
}

func (a *Archiver) Run(ctx context.Context) {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	lastRun := make(map[string]time.Time) // key -> last run

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			a.mu.RLock()
			for k, w := range a.watches {
				if time.Since(lastRun[k]) < w.Every { continue }
				lastRun[k] = time.Now()
				go a.fetchOnce(w)
			}
			a.mu.RUnlock()
		}
	}
}

func (a *Archiver) fetchOnce(w Watch) {
	src, ok := a.sources[w.Source]
	if !ok { return }
	th, err := src.FetchThread(w.Board, w.ThreadID)
	if err != nil { log.Printf("archiver fetch error: %v", err); return }
	if err := a.store.SaveThread(*th); err != nil { log.Printf("archiver save error: %v", err) }
}
