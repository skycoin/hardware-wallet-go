package interfaces

import "io"

import "github.com/skycoin/hardware-wallet-go/src/device-wallet/wire"

// DeviceType type of device: emulator or usb
type DeviceType int32

const (
	// DeviceTypeEmulator use emulator
	DeviceTypeEmulator DeviceType = 1
	// DeviceTypeEmulatorStr string to represent DeviceTypeEmulator
	DeviceTypeEmulatorStr string = "EMULATOR"
	// DeviceTypeUsb use usb
	DeviceTypeUSB DeviceType = 2
	// DeviceTypeUSBStr string to represent DeviceTypeUSB
	DeviceTypeUSBStr string = "USB"
	// DeviceTypeInvalid
	DeviceTypeInvalid DeviceType = 3
)

type DeviceDriver interface {
	SendToDevice(dev io.ReadWriteCloser, chunks [][64]byte) (wire.Message, error)
	SendToDeviceNoAnswer(dev io.ReadWriteCloser, chunks [][64]byte) error
	DeviceType() DeviceType
}