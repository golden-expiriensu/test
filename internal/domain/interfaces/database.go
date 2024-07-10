package interfaces

type Database interface {
	Get([]byte) (string, error)
}
