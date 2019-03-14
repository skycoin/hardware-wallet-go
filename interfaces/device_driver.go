package interfaces

import (
	"io"

	"github.com/skycoin/hardware-wallet-go/src/device-wallet/wire"
)

// DeviceType type of device: emulator or usb
type DeviceType int32

func (dt DeviceType) String() string {
	switch dt {
	case DeviceTypeEmulator:
		return "EMULATOR"
	case DeviceTypeUSB:
		return "USB"
	default:
		return "Invalid"
	}
}

const (
	// DeviceTypeEmulator use emulator
	DeviceTypeEmulator = iota + 1
	// DeviceTypeUsb use usb
	DeviceTypeUSB
	// DeviceTypeInvalid not valid value
	DeviceTypeInvalid
)

type DeviceDriver interface {
	SendToDevice(dev io.ReadWriteCloser, chunks [][64]byte) (wire.Message, error)
	SendToDeviceNoAnswer(dev io.ReadWriteCloser, chunks [][64]byte) error
	GetDevice() (io.ReadWriteCloser, error)
	DeviceType() DeviceType
}
