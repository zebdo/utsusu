package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/zebdo/utsusu/internal/chans"
	"github.com/zebdo/utsusu/internal/core"
	"github.com/zebdo/utsusu/internal/storage"
)

type Server struct {
	r *gin.Engine
	store storage.Storage
	sources map[string]chans.ChanSource
	arch *core.Archiver
	adminToken string
}

func New(store storage.Storage, sources map[string]chans.ChanSource, arch *core.Archiver, adminToken string) *Server {
	r := gin.Default()
	s := &Server{ r: r, store: store, sources: sources, arch: arch, adminToken: adminToken }
	s.routes()
	return s
}

func (s *Server) Run(addr string) error { return s.r.Run(addr) }

func (s *Server) routes() {
	s.r.GET("/health", func(c *gin.Context){ c.JSON(200, gin.H{"ok": true}) })
	s.r.GET("/api/threads/:board", s.getThreads)
	s.r.GET("/api/thread/:id", s.getThread)

	adm := s.r.Group("/", AdminAuth(s.adminToken))
	adm.POST("/api/fetch/:source/:board/:threadID", s.fetchThread)
	adm.POST("/api/watch/:source/:board/:threadID", s.addWatch)
	adm.GET("/api/watches", s.listWatches)
	adm.POST("/api/scan/:source/:board", s.scanBoard)
}

func (s *Server) getThreads(c *gin.Context) {
	board := c.Param("board")
	threads, err := s.store.ListThreads(board)
	if err != nil { c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}); return }
	c.JSON(200, threads)
}

func (s *Server) getThread(c *gin.Context) {
	id := c.Param("id")
	th, err := s.store.GetThread(id)
	if err != nil { c.JSON(404, gin.H{"error": "not found"}); return }
	c.JSON(200, th)
}

func (s *Server) fetchThread(c *gin.Context) {
	src := c.Param("source")
	board := c.Param("board")
	threadID := c.Param("threadID")

	source, ok := s.sources[src]
	if !ok { c.JSON(400, gin.H{"error": "unknown source"}); return }

	th, err := source.FetchThread(board, threadID)
	if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }

	if err := s.store.SaveThread(*th); err != nil {
		c.JSON(500, gin.H{"error": err.Error()}); return
	}
	c.JSON(200, gin.H{"ok": true, "saved_id": th.ID})
}

func (s *Server) addWatch(c *gin.Context) {
	src := c.Param("source")
	board := c.Param("board")
	threadID := c.Param("threadID")
	everyStr := c.Query("every") // e.g. 30s, 1m
	if everyStr == "" { everyStr = "30s" }
	every, err := time.ParseDuration(everyStr)
	if err != nil { c.JSON(400, gin.H{"error": "bad duration"}); return }
	s.arch.AddWatch(core.Watch{ Source: src, Board: board, ThreadID: threadID, Every: every })
	c.JSON(200, gin.H{"ok": true})
}

func (s *Server) listWatches(c *gin.Context) {
	w := s.arch.ListWatches()
	c.JSON(200, w)
}

// scanBoard -- fetches catalog, saves threads (and optionally adds watches)
// query params:
//   limit=int  (max threads to ingest)
//   watch=bool (if true, add watch with 'every' param)
func (s *Server) scanBoard(c *gin.Context) {
	src := c.Param("source")
	board := c.Param("board")
	limitStr := c.Query("limit")
	limit := 0
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil { limit = v }
	}
	watchFlag := c.Query("watch") == "true"
	everyStr := c.Query("every")
	if everyStr == "" { everyStr = "30s" }
	every, _ := time.ParseDuration(everyStr)

	source, ok := s.sources[src]
	if !ok { c.JSON(400, gin.H{"error": "unknown source"}); return }

	threads, err := source.FetchBoard(board, limit)
	if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }

	saved := 0
	for _, t := range threads {
		// fetch full thread (some sources may not include posts in catalog)
		if full, err := source.FetchThread(board, t.ID); err == nil {
			if err := s.store.SaveThread(*full); err == nil { saved++ }
		}
		if watchFlag {
			s.arch.AddWatch(core.Watch{ Source: src, Board: board, ThreadID: t.ID, Every: every })
		}
	}
	c.JSON(200, gin.H{"ok": true, "scanned": len(threads), "saved": saved})
}
