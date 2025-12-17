package data

type MasterQ interface {
	New() MasterQ
	User() Userdb
}
