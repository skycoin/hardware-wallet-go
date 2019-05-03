package devicewallet

type BitEncodedFlags interface {
	Marshal(v interface{}) (uint64, error)
	Unmarshal() error
}