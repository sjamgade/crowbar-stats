package storage

type Service interface {
	Startrun(string) (string, error)
	Stoprun(string, []byte) (string, error)
	Close() error
}
