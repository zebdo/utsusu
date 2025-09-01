package chans

import (
"encoding/json"
"fmt"
"io"
"net/http"
"strconv"
"time"

"github.com/zebdo/utsusu/internal/core"
)

type FourChanConfig struct {
	UserAgent string
}

type FourChan struct {
	client *http.Client
	ua     string
}

func NewFourChan(cfg FourChanConfig) *FourChan {
	if cfg.UserAgent == "" { cfg.UserAgent = "GoChan/0.3" }
		return &FourChan{
			client: &http.Client{ Timeout: 15 * time.Second },
			ua: cfg.UserAgent,
		}
}

func (fc *FourChan) Name() string { return "4chan" }

type fcPost struct {
	No       int64   `json:"no"`
	Resto    int64   `json:"resto"`
	Name     string  `json:"name"`
	Com      string  `json:"com"` // HTML from API
	Time     int64   `json:"time"`
	Tim      *int64  `json:"tim,omitempty"` // image timestamp id
	Ext      string  `json:"ext,omitempty"`
	Filename string  `json:"filename,omitempty"`
	Sticky   *int    `json:"sticky,omitempty"`
	Closed   *int    `json:"closed,omitempty"`
	MD5      string  `json:"md5,omitempty"`
}

type fcThreadResp struct {
	Posts []fcPost `json:"posts"`
}

type fcCatalogPage struct {
	Page    int `json:"page"`
	Threads []struct {
		No           int64  `json:"no"`
		LastModified int64  `json:"last_modified"`
		Sticky       *int   `json:"sticky,omitempty"`
		Closed       *int   `json:"closed,omitempty"`
		Sub          string `json:"sub,omitempty"`
		Com          string `json:"com,omitempty"`
	} `json:"threads"`
}

func (fc *FourChan) get(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", fc.ua)
	resp, err := fc.client.Do(req)
	if err != nil { return nil, err }
		defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s: %s", resp.Status, string(b))
	}
	return io.ReadAll(resp.Body)
}

func (fc *FourChan) FetchThread(board, threadID string) (*core.Thread, error) {
	url := fmt.Sprintf("https://a.4cdn.org/%s/thread/%s.json", board, threadID)
	b, err := fc.get(url)
	if err != nil { return nil, err }

	var tr fcThreadResp
	if err := json.Unmarshal(b, &tr); err != nil { return nil, err }

	th := &core.Thread{ ID: threadID, Board: board }
	if len(tr.Posts) > 0 {
		if tr.Posts[0].Sticky != nil && *tr.Posts[0].Sticky == 1 { th.Sticky = true }
		if tr.Posts[0].Closed != nil && *tr.Posts[0].Closed == 1 { th.Closed = true }
	}
	th.Posts = make([]core.Post, 0, len(tr.Posts))

	for _, p := range tr.Posts {
		post := core.Post{
			ID: strconv.FormatInt(p.No, 10),
			Author: p.Name,
			Content: p.Com,
			Timestamp: time.Unix(p.Time, 0),
			Metadata: map[string]any{},
		}
		if p.Tim != nil && p.Ext != "" {
			imgURL := fmt.Sprintf("https://i.4cdn.org/%s/%d%s", board, *p.Tim, p.Ext)
			thumb := fmt.Sprintf("https://i.4cdn.org/%s/%ds.jpg", board, *p.Tim)
			post.Images = append(post.Images, core.Image{ URL: imgURL, Thumbnail: thumb, MD5: p.MD5 })
		}
		th.Posts = append(th.Posts, post)
	}
	return th, nil
}

func (fc *FourChan) FetchBoard(board string, limit int) ([]core.Thread, error) {
	url := fmt.Sprintf("https://a.4cdn.org/%s/catalog.json", board)
	b, err := fc.get(url)
	if err != nil { return nil, err }
	var pages []fcCatalogPage
	if err := json.Unmarshal(b, &pages); err != nil { return nil, err }
	threads := make([]core.Thread, 0)
	count := 0
	for _, pg := range pages {
		for _, t := range pg.Threads {
			th := core.Thread{ ID: strconv.FormatInt(t.No, 10), Board: board }
			if t.Sticky != nil && *t.Sticky == 1 { th.Sticky = true }
			if t.Closed != nil && *t.Closed == 1 { th.Closed = true }
			if th.Metadata == nil { th.Metadata = map[string]any{} }
			th.Metadata["last_modified"] = t.LastModified
			threads = append(threads, th)
			count++
			if limit > 0 && count >= limit { return threads, nil }
		}
	}
	return threads, nil
}
