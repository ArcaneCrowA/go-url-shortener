package service

type Repo interface {
	InsertRecord(code, site string) error
	GetRecord(code string) (string, error)
}
