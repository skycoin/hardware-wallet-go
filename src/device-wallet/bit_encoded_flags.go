package devicewallet

type BitEncodedFlags interface {
	Marshal(v interface{}) (uint64, error)
	Unmarshal() error
}

func bitStatusInByte(data, bitPos uint8) bool {
	return (data & (uint8)(1 << bitPos)) != 0
}

func setBitInByte(data *uint8, val bool, bitPos uint8) {
	mask := (uint8)(1 << bitPos)
	if val {
		*data |= mask
	} else {
		*data &= mask ^ 255
	}
}