package data

type Userdb interface {
	Insert(User) (*User, error)
}

type User struct {
	ID           int64  `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"hashed_password"`
}
