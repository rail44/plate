package tables

// User represents the user table
type User struct{}

func (User) TableName() string { return "user" }

// Profile represents the profile table  
type Profile struct{}

func (Profile) TableName() string { return "profile" }

// Order represents the order table (for future extension)
type Order struct{}

func (Order) TableName() string { return "order" }