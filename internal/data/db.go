package data

type Userdb interface {
	Insert(User) (*User, error)
	Get(username string) (*User, error)
	GetByUsername(username string) (*User, error)
}

type User struct {
	ID           int64  `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"hashed_password"`
}

type Addressdb interface {
	Insert(Address) error
}

type Address struct {
	ID      int64  `db:"id"`
	UserID  int64  `db:"user_id"`
	Address string `db:"address"`
}
