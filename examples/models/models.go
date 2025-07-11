package models

import "time"

// User represents a user in the system
type User struct {
	ID        string    `spanner:"id" spannerType:"STRING"`
	Name      string    `spanner:"name" spannerType:"STRING"`
	Email     string    `spanner:"email" spannerType:"STRING"`
	CreatedAt time.Time `spanner:"created_at" spannerType:"TIMESTAMP"`
}

// Post represents a blog post
type Post struct {
	ID        string    `spanner:"id" spannerType:"STRING"`
	UserID    string    `spanner:"user_id" spannerType:"STRING"`
	Title     string    `spanner:"title" spannerType:"STRING"`
	Content   string    `spanner:"content" spannerType:"STRING"`
	CreatedAt time.Time `spanner:"created_at" spannerType:"TIMESTAMP"`
}

// Tag represents a tag for posts
type Tag struct {
	ID   string `spanner:"id" spannerType:"STRING"`
	Name string `spanner:"name" spannerType:"STRING"`
}

// PostTag represents the junction table for post-tag relationships
type PostTag struct {
	PostID    string    `spanner:"post_id" spannerType:"STRING"`
	TagID     string    `spanner:"tag_id" spannerType:"STRING"`
	CreatedAt time.Time `spanner:"created_at" spannerType:"TIMESTAMP"`
}
