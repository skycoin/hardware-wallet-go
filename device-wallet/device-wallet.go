package devicewallet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"time"

	"github.com/skycoin/hardware-wallet-go/device-wallet/usb"
	"github.com/skycoin/hardware-wallet-go/device-wallet/wire"

	proto "github.com/golang/protobuf/proto"
	messages "github.com/skycoin/hardware-wallet-go/device-wallet/messages"
)

// DeviceType type of device: emulated or usb
type DeviceType int32

const (
	// DeviceTypeEmulator use emulator
	DeviceTypeEmulator DeviceType = 1
	// DeviceTypeUsb use usb
	DeviceTypeUsb DeviceType = 2
)

func getEmulatorDevice() (net.Conn, error) {
	return net.Dial("udp", "127.0.0.1:21324")
}

func getUsbDevice() (usb.Device, error) {
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

func sendToDeviceNoAnswer(dev io.ReadWriteCloser, chunks [][64]byte) error {
	for _, element := range chunks {
		_, err := dev.Write(element[:])
		if err != nil {
			return err
		}
	}
	return nil
}
func sendToDevice(dev io.ReadWriteCloser, chunks [][64]byte) (wire.Message, error) {
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

func binaryWrite(message io.Writer, data interface{}) {
	err := binary.Write(message, binary.BigEndian, data)
	if err != nil {
		log.Print(err.Error())
	}
}

func makeTrezorMessage(data []byte, msgID messages.MessageType) [][64]byte {
	message := new(bytes.Buffer)
	binaryWrite(message, []byte("##"))
	binaryWrite(message, uint16(msgID))
	binaryWrite(message, uint32(len(data)))
	binaryWrite(message, []byte("\n"))
	if len(data) > 0 {
		binaryWrite(message, data[1:])
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

func getDevice(deviceType DeviceType) (io.ReadWriteCloser, error) {
	var dev io.ReadWriteCloser
	var err error
	switch deviceType {
	case DeviceTypeEmulator:
		dev, err = getEmulatorDevice()
	case DeviceTypeUsb:
		dev, err = getUsbDevice()
	}
	if dev == nil && err == nil {
		err = errors.New("No device connected")
	}
	return dev, err
}

// DeviceCheckMessageSignature Check a message signature matches the given address.
func DeviceCheckMessageSignature(deviceType DeviceType, message string, signature string, address string) (uint16, []byte) {

	dev, err := getDevice(deviceType)
	if err != nil {
		log.Panicf(err.Error())
		return 0, make([]byte, 0)
	}
	defer dev.Close()

	// Send CheckMessageSignature

	skycoinCheckMessageSignature := &messages.SkycoinCheckMessageSignature{
		Address:   proto.String(address),
		Message:   proto.String(message),
		Signature: proto.String(signature),
	}

	data, _ := proto.Marshal(skycoinCheckMessageSignature)
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_SkycoinCheckMessageSignature)
	msg, err := sendToDevice(dev, chunks)
	if err != nil {
		log.Print(err.Error())
		return msg.Kind, msg.Data
	}
	log.Printf("Success %d! address that issued the signature is: %s\n", msg.Kind, msg.Data)
	return msg.Kind, msg.Data
}

// MessageCancel prepare Cancel request
func MessageCancel() [][64]byte {
	msg := &messages.Cancel{}
	data, _ := proto.Marshal(msg)
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_Cancel)
	return chunks
}

// MessageButtonAck send this message (before user action) when the device expects the user to push a button
func MessageButtonAck() [][64]byte {
	buttonAck := &messages.ButtonAck{}
	data, _ := proto.Marshal(buttonAck)
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_ButtonAck)
	return chunks
}

// MessagePassphraseAck send this message when the device expects receiving a Passphrase
func MessagePassphraseAck(passphrase string) [][64]byte {
	msg := &messages.PassphraseAck{
		Passphrase: proto.String(passphrase),
	}
	data, _ := proto.Marshal(msg)
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_PassphraseAck)
	return chunks
}

// MessageWordAck send this message between each word of the seed (before user action) during device backup
func MessageWordAck(word string) [][64]byte {
	wordAck := &messages.WordAck{
		Word: proto.String(word),
	}
	data, _ := proto.Marshal(wordAck)
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_WordAck)
	return chunks
}

// DeviceButtonAck when the device is waiting for the user to press a button
// the PC need to acknowledge, showing it knows we are waiting for a user action
func DeviceButtonAck(deviceType DeviceType) wire.Message {
	dev, err := getDevice(deviceType)
	if err != nil {
		log.Panicf(err.Error())
	}
	defer dev.Close()
	return deviceButtonAck(dev)
}

func deviceButtonAck(dev io.ReadWriteCloser) wire.Message {
	var msg wire.Message
	// Send ButtonAck
	chunks := MessageButtonAck()
	err := sendToDeviceNoAnswer(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
	}

	_, err = msg.ReadFrom(dev)
	time.Sleep(1 * time.Second)
	if err != nil {
		log.Panicf(err.Error())
	}
	return msg
}

// DevicePassphraseAck send this message when the device is waiting for the user to input a passphrase
func DevicePassphraseAck(deviceType DeviceType, passphrase string) (uint16, []byte) {
	dev, err := getDevice(deviceType)
	if err != nil {
		log.Panicf(err.Error())
	}
	defer dev.Close()
	msg := devicePassphraseAck(dev, passphrase)
	return msg.Kind, msg.Data
}

func devicePassphraseAck(dev io.ReadWriteCloser, passphrase string) wire.Message {
	var msg wire.Message
	chunks := MessagePassphraseAck(passphrase)
	msg, err := sendToDevice(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
	}
	return msg
}

// DeviceCancel send Cancel request
func DeviceCancel(deviceType DeviceType) {
	dev, err := getDevice(deviceType)
	if err != nil {
		log.Panicf(err.Error())
	}
	defer dev.Close()

	chunks := MessageCancel()
	msg, err := sendToDevice(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
	}
	log.Println(DecodeSuccessOrFailMsg(msg.Kind, msg.Data))
}

// DeviceFirmwareUpload Updates device's firmware
func DeviceFirmwareUpload(payload []byte, hash [32]byte) {
	dev, err := getDevice(DeviceTypeUsb)
	if err != nil {
		log.Panicf(err.Error())
		return
	}
	defer dev.Close()

	initialize(dev)

	log.Printf("Length of firmware %d", uint32(len(payload)))
	deviceFirmwareErase := &messages.FirmwareErase{
		Length: proto.Uint32(uint32(len(payload))),
	}

	erasedata, err := proto.Marshal(deviceFirmwareErase)
	if err != nil {
		log.Panicf(err.Error())
		return
	}
	// log.Printf("Data: %s\n", data)
	chunks := makeTrezorMessage(erasedata, messages.MessageType_MessageType_FirmwareErase)

	erasemsg, _ := sendToDevice(dev, chunks)
	log.Printf("Success %d! FirmwareErase %s\n", erasemsg.Kind, erasemsg.Data)

	log.Printf("Hash: %x\n", hash)
	deviceFirmwareUpload := &messages.FirmwareUpload{
		Payload: payload,
		Hash:    hash[:],
	}

	uploaddata, err := proto.Marshal(deviceFirmwareUpload)
	if err != nil {
		log.Panicf(err.Error())
		return
	}
	chunks = makeTrezorMessage(uploaddata, messages.MessageType_MessageType_FirmwareUpload)

	uploadmsg, err := sendToDevice(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
		return
	}
	log.Printf("Success %d! FirmwareUpload %s\n", uploadmsg.Kind, uploadmsg.Data)

	// Send ButtonAck
	chunks = MessageButtonAck()
	err = sendToDeviceNoAnswer(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
		return
	}
}

// DeviceSetMnemonic Configure the device with a mnemonic.
func DeviceSetMnemonic(deviceType DeviceType, mnemonic string) {

	dev, err := getDevice(deviceType)
	if err != nil {
		log.Panicf(err.Error())
		return
	}
	defer dev.Close()

	// Send SetMnemonic

	skycoinSetMnemonic := &messages.SetMnemonic{
		Mnemonic: proto.String(mnemonic),
	}

	data, _ := proto.Marshal(skycoinSetMnemonic)
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_SetMnemonic)

	msg, err := sendToDevice(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
		return
	}

	log.Printf("Success %d! Mnemonic %s\n", msg.Kind, msg.Data)
	if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
		msg = deviceButtonAck(dev)
	}
	log.Println(DecodeSuccessOrFailMsg(msg.Kind, msg.Data))
}

// DeviceGetVersion Ask the firmware version
func DeviceGetVersion(deviceType DeviceType) string {

	dev, err := getDevice(deviceType)
	if err != nil {
		log.Panicf(err.Error())
		return ""
	}
	defer dev.Close()

	skycoinGetVersion := &messages.GetVersion{}

	data, _ := proto.Marshal(skycoinGetVersion)
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_GetVersion)

	msg, err := sendToDevice(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
		return ""
	}
	return DecodeSuccessOrFailMsg(msg.Kind, msg.Data)
}

// DeviceGenerateMnemonic Ask the device to generate a mnemonic and configure itself with it.
func DeviceGenerateMnemonic(deviceType DeviceType, usePassphrase bool) {

	dev, err := getDevice(deviceType)
	if err != nil {
		log.Panicf(err.Error())
		return
	}
	defer dev.Close()

	skycoinGenerateMnemonic := &messages.GenerateMnemonic{
		PassphraseProtection: proto.Bool(usePassphrase),
	}

	data, _ := proto.Marshal(skycoinGenerateMnemonic)
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_GenerateMnemonic)

	msg, err := sendToDevice(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
		return
	}

	if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
		msg = deviceButtonAck(dev)
	}
	log.Println(DecodeSuccessOrFailMsg(msg.Kind, msg.Data))
}

func DecodeSuccessOrFailMsg(kind uint16, data []byte) string {

	if kind == uint16(messages.MessageType_MessageType_Success) {
		return DecodeSuccessMsg(kind, data)
	}
	if kind == uint16(messages.MessageType_MessageType_Failure) {
		return DecodeFailMsg(kind, data)
	}
	log.Printf("Calling DecodeSuccessOrFailMsg on message kind %d", kind)
	return ""
}

// DecodeSuccessMsg convert byte data into string containing the success message returned by the device
func DecodeSuccessMsg(kind uint16, data []byte) string {
	if kind == uint16(messages.MessageType_MessageType_Success) {
		success := &messages.Success{}
		err := proto.Unmarshal(data, success)
		if err != nil {
			log.Panicf("unmarshaling error: %s\n", err.Error())
			return ""
		}
		return success.GetMessage()
	}
	log.Panicf("Calling DecodeSuccessMsg with message type %d", kind)
	return ""
}

// DecodeFailMsg convert byte data into string containing the failure returned by the device
func DecodeFailMsg(kind uint16, data []byte) string {
	if kind == uint16(messages.MessageType_MessageType_Failure) {
		failure := &messages.Failure{}
		err := proto.Unmarshal(data, failure)
		if err != nil {
			log.Panicf("unmarshaling error: %s\n", err.Error())
			return ""
		}
		return failure.GetMessage()
	}
	log.Panicf("Calling DecodeFailMsg with message type %d", kind)
	return ""
}

// DecodeResponseSkycoinAddress convert byte data into list of addresses, meant to be used after DevicePinMatrixAck
func DecodeResponseSkycoinAddress(kind uint16, data []byte) (uint16, []string) {
	log.Printf("%x\n", data)
	if kind == uint16(messages.MessageType_MessageType_ResponseSkycoinAddress) {
		responseSkycoinAddress := &messages.ResponseSkycoinAddress{}
		err := proto.Unmarshal(data, responseSkycoinAddress)
		if err != nil {
			log.Panicf("unmarshaling error: %s\n", err.Error())
			return kind, make([]string, 0)
		}
		return kind, responseSkycoinAddress.GetAddresses()
	}
	log.Panic("Calling DecodeResponseSkycoinAddress with wrong message type")
	return kind, make([]string, 0)
}

// DecodeResponseSkycoinSignMessage convert byte data into signed message, meant to be used after DevicePinMatrixAck
func DecodeResponseSkycoinSignMessage(kind uint16, data []byte) (uint16, string) {
	if kind == uint16(messages.MessageType_MessageType_ResponseSkycoinSignMessage) {
		responseSkycoinSignMessage := &messages.ResponseSkycoinSignMessage{}
		err := proto.Unmarshal(data, responseSkycoinSignMessage)
		if err != nil {
			log.Panicf("unmarshaling error: %s\n", err.Error())
			return kind, ""
		}
		return kind, responseSkycoinSignMessage.GetSignedMessage()
	}
	log.Panic("Calling DecodeResponseeSkycoinSignMessage with wrong message type")
	return kind, ""
}

// DeviceAddressGen Ask the device to generate an address
func DeviceAddressGen(deviceType DeviceType, addressN int, startIndex int, confirmAddress bool) (uint16, []byte) {

	dev, err := getDevice(deviceType)
	if err != nil {
		log.Panicf(err.Error())
	}
	defer dev.Close()
	skycoinAddress := &messages.SkycoinAddress{
		AddressN:       proto.Uint32(uint32(addressN)),
		ConfirmAddress: proto.Bool(confirmAddress),
		StartIndex:     proto.Uint32(uint32(startIndex)),
	}
	data, _ := proto.Marshal(skycoinAddress)

	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_SkycoinAddress)

	msg, err := sendToDevice(dev, chunks)
	if err != nil {
		log.Panicf("sendToDevice error: %s\n", err.Error())
	}
	return msg.Kind, msg.Data
}

// DeviceSignMessage Ask the device to sign a message using the secret key at given index.
func DeviceSignMessage(deviceType DeviceType, addressN int, message string) (uint16, []byte) {

	dev, err := getDevice(deviceType)
	if err != nil {
		log.Panicf(err.Error())
	}
	defer dev.Close()

	skycoinSignMessage := &messages.SkycoinSignMessage{
		AddressN: proto.Uint32(uint32(addressN)),
		Message:  proto.String(message),
	}

	data, _ := proto.Marshal(skycoinSignMessage)

	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_SkycoinSignMessage)

	msg, err := sendToDevice(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
	}
	return msg.Kind, msg.Data
}

// DeviceConnected check if a device is connected
func DeviceConnected(deviceType DeviceType) bool {
	dev, err := getDevice(deviceType)
	if dev == nil {
		return false
	}
	defer dev.Close()
	if err != nil {
		return false
	}
	msgRaw := &messages.Ping{}
	data, err := proto.Marshal(msgRaw)
	if err != nil {
		log.Print(err.Error())
	}
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_Ping)
	for _, element := range chunks {
		_, err = dev.Write(element[:])
		if err != nil {
			return false
		}
	}
	var msg wire.Message
	_, err = msg.ReadFrom(dev)
	if err != nil {
		return false
	}
	return msg.Kind == uint16(messages.MessageType_MessageType_Success)
}

// Initialize send an init request to the device
func initialize(dev io.ReadWriteCloser) {
	var chunks [][64]byte

	initialize := &messages.Initialize{}
	data, _ := proto.Marshal(initialize)
	chunks = makeTrezorMessage(data, messages.MessageType_MessageType_Initialize)
	_, err := sendToDevice(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
		return
	}
}

// DeviceGetFeatures send Features message to the device
func DeviceGetFeatures(deviceType DeviceType) {
	dev, err := getDevice(deviceType)
	if err != nil {
		log.Panicf(err.Error())
		return
	}
	defer dev.Close()

	featureMsg := &messages.GetFeatures{}
	data, _ := proto.Marshal(featureMsg)
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_GetFeatures)
	msg, err := sendToDevice(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
	}
	if msg.Kind == uint16(messages.MessageType_MessageType_Failure) || msg.Kind == uint16(messages.MessageType_MessageType_Success) {
		log.Printf("Received message kind: %d %s\n", msg.Kind, DecodeSuccessOrFailMsg(msg.Kind, msg.Data))
		return
	}
	features := &messages.Features{}
	err = proto.Unmarshal(msg.Data, features)
	if err != nil {
		log.Panicf("unmarshaling error: %s\n", err.Error())
	}
	log.Printf(`Vendor: %s
MajorVersion: %d
MinorVersion: %d
PatchVersion: %d
BootloaderMode: %t
DeviceId: %x
PinProtection: %t
PassphraseProtection: %t
Language: %s
Label: %s
Initialized: %t
BootloaderHash: %x
PinCached: %t
PassphraseCached: %t
FirmwarePresent: %t
NeedsBackup: %t
Model: %s
FwMajor: %d
FwMinor: %d
FwPatch: %d
FwVendor: %s
FwVendorKeys: %s
UnfinishedBackup: %t`,
		features.GetVendor(),
		features.GetMajorVersion(),
		features.GetMinorVersion(),
		features.GetPatchVersion(),
		features.GetBootloaderMode(),
		features.GetDeviceId(),
		features.GetPinProtection(),
		features.GetPassphraseProtection(),
		features.GetLanguage(),
		features.GetLabel(),
		features.GetInitialized(),
		features.GetBootloaderHash(),
		features.GetPinCached(),
		features.GetPassphraseCached(),
		features.GetFirmwarePresent(),
		features.GetNeedsBackup(),
		features.GetModel(),
		features.GetFwMajor(),
		features.GetFwMinor(),
		features.GetFwPatch(),
		features.GetFwVendor(),
		features.GetFwVendorKeys(),
		features.GetUnfinishedBackup())
}

// BackupDevice ask the device to perform the seed backup
func BackupDevice(deviceType DeviceType) wire.Message {
	dev, err := getDevice(deviceType)
	if err != nil {
		log.Panicf(err.Error())
	}
	defer dev.Close()
	var msg wire.Message
	var chunks [][64]byte
	initialize(dev)

	backupDevice := &messages.BackupDevice{}
	data, _ := proto.Marshal(backupDevice)
	chunks = makeTrezorMessage(data, messages.MessageType_MessageType_BackupDevice)
	msg, err = sendToDevice(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
	}

	for msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
		msg = deviceButtonAck(dev)
	}
	return msg
}

// DeviceWordAck send a word to the device during device "recovery procedure"
func DeviceWordAck(deviceType DeviceType, word string) wire.Message {
	dev, err := getDevice(deviceType)
	if err != nil {
		log.Panicf(err.Error())
	}
	defer dev.Close()
	chunks := MessageWordAck(word)
	msg, err := sendToDevice(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
	}
	return msg
}

// RecoveryDevice ask the device to perform the seed backup
func RecoveryDevice(deviceType DeviceType, usePassphrase bool) wire.Message {
	dev, err := getDevice(deviceType)
	if err != nil {
		log.Panicf(err.Error())
	}
	defer dev.Close()
	var msg wire.Message
	var chunks [][64]byte

	log.Printf("Using passphrase %t\n", usePassphrase)

	recoveryDevice := &messages.RecoveryDevice{
		DryRun:               proto.Bool(false),
		EnforceWordlist:      proto.Bool(true),
		WordCount:            proto.Uint32(12),
		PassphraseProtection: proto.Bool(usePassphrase),
	}
	data, _ := proto.Marshal(recoveryDevice)
	chunks = makeTrezorMessage(data, messages.MessageType_MessageType_RecoveryDevice)
	msg, err = sendToDevice(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
	}
	log.Printf("Recovery device %d! Answer is: %s\n", msg.Kind, msg.Data)

	if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
		msg = deviceButtonAck(dev)
	}
	return msg
}

// WipeDevice wipes out device configuration
func WipeDevice(deviceType DeviceType) {
	dev, err := getDevice(deviceType)
	if err != nil {
		log.Panicf(err.Error())
		return
	}
	defer dev.Close()
	var msg wire.Message
	var chunks [][64]byte

	initialize(dev)

	wipeDevice := &messages.WipeDevice{}
	data, _ := proto.Marshal(wipeDevice)
	chunks = makeTrezorMessage(data, messages.MessageType_MessageType_WipeDevice)
	msg, err = sendToDevice(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
		return
	}
	log.Printf("Wipe device %d! Answer is: %x\n", msg.Kind, msg.Data)

	if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
		msg = deviceButtonAck(dev)
	}
	log.Println(DecodeSuccessOrFailMsg(msg.Kind, msg.Data))

	initialize(dev)
}

// DeviceChangePin changes device's PIN code
// The message that is sent contains an encoded form of the PIN.
// The digits of the PIN are displayed in a 3x3 matrix on the Trezor,
// and the message that is sent back is a string containing the positions
// of the digits on that matrix. Below is the mapping between positions
// and characters to be sent:
// 7 8 9
// 4 5 6
// 1 2 3
// For example, if the numbers are laid out in this way on the Trezor,
// 3 1 5
// 7 8 4
// 9 6 2
// To set the PIN "12345", the positions are:
// top, bottom-right, top-left, right, top-right
// so you must send "83769".
func DeviceChangePin(deviceType DeviceType) (uint16, []byte) {
	dev, err := getDevice(deviceType)
	if err != nil {
		log.Panicf(err.Error())
		return 0, make([]byte, 0)
	}
	defer dev.Close()

	changePin := &messages.ChangePin{}
	data, _ := proto.Marshal(changePin)
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_ChangePin)
	msg, err := sendToDevice(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
		return msg.Kind, msg.Data
	}
	// Acknowledge that a button has been pressed
	if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
		msg = deviceButtonAck(dev)
	}
	return msg.Kind, msg.Data
}

// DevicePinMatrixAck during PIN code setting use this message to send user input to device
func DevicePinMatrixAck(deviceType DeviceType, p string) (uint16, []byte) {
	time.Sleep(1 * time.Second)
	dev, err := getDevice(deviceType)
	if err != nil {
		log.Panicf(err.Error())
		return 0, make([]byte, 0)
	}
	defer dev.Close()
	var msg wire.Message
	log.Printf("Setting pin: %s\n", p)
	pinAck := &messages.PinMatrixAck{
		Pin: proto.String(p),
	}
	data, _ := proto.Marshal(pinAck)

	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_PinMatrixAck)
	msg, err = sendToDevice(dev, chunks)
	if err != nil {
		log.Panicf(err.Error())
		return msg.Kind, msg.Data
	}
	if msg.Kind != uint16(messages.MessageType_MessageType_PinMatrixAck) {
		log.Printf("MessagePinMatrixAck Answer is: %d / %s\n", msg.Kind, DecodeSuccessOrFailMsg(msg.Kind, msg.Data))
	}
	return msg.Kind, msg.Data
}
