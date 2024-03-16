package dto

import "time"

// Article represents a blog article
type Article struct {
	ID       int
	Title    string
	Author   string
	Link     string
	PostTime time.Time
	Votes    int
	Group    string
}
