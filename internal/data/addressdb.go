package data

type Addressdb interface {
	Insert(Address) error
	Select(userID int64) ([]Address, error)
	GetByAddress(address string) (*Address, error)
	Get() (*Address, error)
	GetByAddressUserID(address string, userID int64) (*Address, error)
}

type Address struct {
	ID      int64  `db:"id"`
	UserID  int64  `db:"user_id"`
	Address string `db:"address"`
}
