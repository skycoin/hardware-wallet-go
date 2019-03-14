package devicewallet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/gogo/protobuf/proto"

	messages "github.com/skycoin/hardware-wallet-go/src/device-wallet/messages/go"
	"github.com/skycoin/hardware-wallet-go/src/device-wallet/usb"
	"github.com/skycoin/hardware-wallet-go/src/device-wallet/wire"

	"io"
	"net"
	"time"
)

type DeviceDriver interface {
	getEmulatorDevice() (net.Conn, error)
	getUsbDevice() (usb.Device, error)
	sendToDeviceNoAnswer(dev io.ReadWriteCloser, chunks [][64]byte) error
	sendToDevice(dev io.ReadWriteCloser, chunks [][64]byte) (wire.Message, error)
	binaryWrite(message io.Writer, data interface{})
	makeTrezorMessage(data []byte, msgID messages.MessageType) [][64]byte
	getDevice() (io.ReadWriteCloser, error)
	getDeviceType() DeviceType
	initialize() error
}

const (
	// DeviceTypeEmulator use emulator
	DeviceTypeEmulator DeviceType = 1
	// DeviceTypeUsb use usb
	DeviceTypeUSB DeviceType = 2
)

type DeviceHelper struct {
	DeviceType
}

func (dh *DeviceHelper) getDeviceType() DeviceType {
	return dh.DeviceType
}

func (dh *DeviceHelper) getEmulatorDevice() (net.Conn, error) {
	return net.Dial("udp", "127.0.0.1:21324")
}

func (dh *DeviceHelper) getUsbDevice() (usb.Device, error) {
	w, err := usb.InitWebUSB()
	if err != nil {
		log.Printf("webusb: %s", err)
		return nil, err
	}
	h, err := usb.InitHIDAPI()
	if err != nil {
		log.Printf("hidapi: %s", err)
		return nil, err
	}
	b := usb.Init(w, h)

	var infos []usb.Info
	infos, err = b.Enumerate()
	if len(infos) <= 0 {
		return nil, err
	}
	tries := 0
	for tries < 3 {
		dev, err := b.Connect(infos[0].Path)
		if err != nil {
			log.Print(err.Error())
			tries++
			time.Sleep(100 * time.Millisecond)
		} else {
			return dev, err
		}
	}
	return nil, err
}

func (dh *DeviceHelper) sendToDeviceNoAnswer(dev io.ReadWriteCloser, chunks [][64]byte) error {
	for _, element := range chunks {
		_, err := dev.Write(element[:])
		if err != nil {
			return err
		}
	}
	return nil
}

func (dh *DeviceHelper) sendToDevice(dev io.ReadWriteCloser, chunks [][64]byte) (wire.Message, error) {
	var msg wire.Message
	for _, element := range chunks {
		_, err := dev.Write(element[:])
		if err != nil {
			return msg, err
		}
	}
	_, err := msg.ReadFrom(dev)
	return msg, err
}

func (dh *DeviceHelper) binaryWrite(message io.Writer, data interface{}) {
	err := binary.Write(message, binary.BigEndian, data)
	if err != nil {
		log.Panic(err)
	}
}

func (dh *DeviceHelper) makeTrezorMessage(data []byte, msgID messages.MessageType) [][64]byte {
	message := new(bytes.Buffer)
	dh.binaryWrite(message, []byte("##"))
	dh.binaryWrite(message, uint16(msgID))
	dh.binaryWrite(message, uint32(len(data)))
	dh.binaryWrite(message, []byte("\n"))
	if len(data) > 0 {
		dh.binaryWrite(message, data[1:])
	}

	messageLen := message.Len()
	var chunks [][64]byte
	i := 0
	for messageLen > 0 {
		var chunk [64]byte
		chunk[0] = '?'
		copy(chunk[1:], message.Bytes()[63*i:63*(i+1)])
		chunks = append(chunks, chunk)
		messageLen -= 63
		i = i + 1
	}
	return chunks
}

func (dh *DeviceHelper) getDevice() (io.ReadWriteCloser, error) {
	var dev io.ReadWriteCloser
	var err error
	switch dh.DeviceType {
	case DeviceTypeEmulator:
		dev, err = dh.getEmulatorDevice()
	case DeviceTypeUSB:
		dev, err = dh.getUsbDevice()
	}
	if dev == nil && err == nil {
		err = errors.New("No device connected")
	}
	return dev, err
}

// Initialize send an init request to the device
func (dh *DeviceHelper) initialize() error {
	dev, err := dh.getDevice()
	if err != nil {
		return err
	}
	defer dev.Close()
	var chunks [][64]byte

	initialize := &messages.Initialize{}
	data, err := proto.Marshal(initialize)
	if err != nil {
		return err
	}

	chunks = dh.makeTrezorMessage(data, messages.MessageType_MessageType_Initialize)
	_, err = dh.sendToDevice(dev, chunks)

	return err
}
