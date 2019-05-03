package devicewallet

import (
	"encoding/binary"
	"encoding/json"
)

type BitEncodedFlags interface {
	Marshal() (uint64, error)
	Unmarshal() error
}

type FirmwareFeatures struct {
	flags uint64
	RequireGetEntropyConfirm bool
	IsGetEntropyEnabled bool
	IsEmulator bool
	FirmwareFeaturesRdpDebugDisabled uint8
	FirmwareFeaturesRdpMemProtect uint8
	FirmwareFeaturesRdpLevel uint8
}

func NewFirmwareFeatures(flags uint64) *FirmwareFeatures {
	return &FirmwareFeatures{flags: flags}
}

func (ff *FirmwareFeatures) Marshal() (uint64, error) {
	ff.flags = 0
	bs := make([]byte, 8)
	setBitInByte(&bs[7], ff.RequireGetEntropyConfirm, 0)
	setBitInByte(&bs[7], ff.IsGetEntropyEnabled, 1)
	setBitInByte(&bs[7], ff.IsEmulator, 2)
	setBitInByte(&bs[7], ff.FirmwareFeaturesRdpLevel == 1, 3)
	setBitInByte(&bs[7], ff.FirmwareFeaturesRdpLevel == 2, 4)
	ff.flags = binary.BigEndian.Uint64(bs)
	return ff.flags, nil
}

func (ff *FirmwareFeatures) Unmarshal() error {
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, ff.flags)
	ff.RequireGetEntropyConfirm = bitStatusInByte(bs[7], 0)
	ff.IsGetEntropyEnabled = bitStatusInByte(bs[7], 1)
	ff.IsEmulator = bitStatusInByte(bs[7], 2)
	setBitInByte(&ff.FirmwareFeaturesRdpLevel, bitStatusInByte(bs[7], 3), 0)
	setBitInByte(&ff.FirmwareFeaturesRdpLevel, bitStatusInByte(bs[7], 4), 1)
	return nil
}

func (ff FirmwareFeatures) String() string {
	b, err := json.Marshal(ff)
	if err != nil {
		return "error rendering FirmwareFeatures " + err.Error()
	}
	return string(b)
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