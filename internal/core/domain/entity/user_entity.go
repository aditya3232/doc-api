package entity

type UserEntity struct {
	ID       int64
	Name     string
	Email    string
	Password string
	Phone    string
	Photo    string
	Address  string
	Token    string
}

type QueryStringUser struct {
	Search    string
	Page      int64
	Limit     int64
	OrderBy   string
	OrderType string
}
