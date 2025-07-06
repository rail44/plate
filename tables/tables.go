package tables

// User represents the user table
type User struct{}

func (User) TableName() string { return "user" }

// Profile represents the profile table  
type Profile struct{}

func (Profile) TableName() string { return "profile" }

// Post represents the post table
type Post struct{}

func (Post) TableName() string { return "post" }