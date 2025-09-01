package structs

import "time"

type Image struct {
	URL       string         `json:"url"`
	Thumbnail string         `json:"thumbnail,omitempty"`
	MD5       string         `json:"md5,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type Post struct {
	ID        string         `json:"id"`
	Author    string         `json:"author,omitempty"`
	Content   string         `json:"content"`
	Timestamp time.Time      `json:"timestamp"`
	Images    []Image        `json:"images,omitempty"`
	Replies   []string       `json:"replies,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type Thread struct {
	ID       string         `json:"id"`
	Board    string         `json:"board"`
	Posts    []Post         `json:"posts"`
	Sticky   bool           `json:"sticky"`
	Closed   bool           `json:"closed"`
	Metadata map[string]any `json:"metadata,omitempty"`
}