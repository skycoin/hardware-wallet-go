package devicewallet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/skycoin/skycoin/src/util/logging"

	"github.com/gogo/protobuf/proto"
	messages "github.com/skycoin/hardware-wallet-go/device-wallet/messages/go"
	"github.com/skycoin/hardware-wallet-go/device-wallet/usb"
	"github.com/skycoin/hardware-wallet-go/device-wallet/wire"
)

// DeviceType type of device: emulated or usb
type DeviceType int32

var (
	log = logging.MustGetLogger("device-wallet")
)

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
func DeviceCheckMessageSignature(deviceType DeviceType, message string, signature string, address string) (wire.Message, error) {
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}
	defer dev.Close()

	// Send CheckMessageSignature

	skycoinCheckMessageSignature := &messages.SkycoinCheckMessageSignature{
		Address:   proto.String(address),
		Message:   proto.String(message),
		Signature: proto.String(signature),
	}

	data, err := proto.Marshal(skycoinCheckMessageSignature)
	if err != nil {
		return wire.Message{}, err
	}
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_SkycoinCheckMessageSignature)
	msg, err := sendToDevice(dev, chunks)
	if err != nil {
		return msg, err
	}
	log.Printf("Success %s! address that issued the signature is: %s\n", messages.MessageType(msg.Kind), msg.Data)
	return msg, nil
}

// MessageCancel prepare Cancel request
func MessageCancel() ([][64]byte, error) {
	msg := &messages.Cancel{}
	data, err := proto.Marshal(msg)
	if err != nil {
		return [][64]byte{}, err
	}
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_Cancel)
	return chunks, nil
}

// MessageButtonAck send this message (before user action) when the device expects the user to push a button
func MessageButtonAck() ([][64]byte, error) {
	buttonAck := &messages.ButtonAck{}
	data, err := proto.Marshal(buttonAck)
	if err != nil {
		return [][64]byte{}, err
	}
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_ButtonAck)
	return chunks, nil
}

// MessagePassphraseAck send this message when the device expects receiving a Passphrase
func MessagePassphraseAck(passphrase string) ([][64]byte, error) {
	msg := &messages.PassphraseAck{
		Passphrase: proto.String(passphrase),
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return [][64]byte{}, err
	}
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_PassphraseAck)
	return chunks, nil
}

// MessageWordAck send this message between each word of the seed (before user action) during device backup
func MessageWordAck(word string) ([][64]byte, error) {
	wordAck := &messages.WordAck{
		Word: proto.String(word),
	}
	data, err := proto.Marshal(wordAck)
	if err != nil {
		return [][64]byte{}, err
	}
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_WordAck)
	return chunks, nil
}

// DeviceButtonAck when the device is waiting for the user to press a button
// the PC need to acknowledge, showing it knows we are waiting for a user action
func DeviceButtonAck(deviceType DeviceType) (wire.Message, error) {
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}
	defer dev.Close()
	return deviceButtonAck(dev)
}

func deviceButtonAck(dev io.ReadWriteCloser) (wire.Message, error) {
	var msg wire.Message
	// Send ButtonAck
	chunks, err := MessageButtonAck()
	if err != nil {
		return msg, err
	}
	err = sendToDeviceNoAnswer(dev, chunks)
	if err != nil {
		return msg, err
	}

	_, err = msg.ReadFrom(dev)
	time.Sleep(1 * time.Second)
	if err != nil {
		return msg, err
	}
	return msg, nil
}

// DevicePassphraseAck send this message when the device is waiting for the user to input a passphrase
func DevicePassphraseAck(deviceType DeviceType, passphrase string) (wire.Message, error) {
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}
	defer dev.Close()

	chunks, err := MessagePassphraseAck(passphrase)
	if err != nil {
		return wire.Message{}, err
	}
	return sendToDevice(dev, chunks)
}

// DeviceCancel send Cancel request
func DeviceCancel(deviceType DeviceType) (wire.Message, error) {
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}
	defer dev.Close()

	chunks, err := MessageCancel()
	if err != nil {
		return wire.Message{}, err
	}

	return sendToDevice(dev, chunks)
}

// DeviceFirmwareUpload Updates device's firmware
func DeviceFirmwareUpload(payload []byte, hash [32]byte) error {
	dev, err := getDevice(DeviceTypeUsb)
	if err != nil {
		return err
	}
	defer dev.Close()

	err = initialize(dev)
	if err != nil {
		return err
	}

	log.Printf("Length of firmware %d", uint32(len(payload)))
	deviceFirmwareErase := &messages.FirmwareErase{
		Length: proto.Uint32(uint32(len(payload))),
	}

	erasedata, err := proto.Marshal(deviceFirmwareErase)
	if err != nil {
		return err
	}

	chunks := makeTrezorMessage(erasedata, messages.MessageType_MessageType_FirmwareErase)
	erasemsg, err := sendToDevice(dev, chunks)
	if err != nil {
		return err
	}
	log.Printf("Success %d! FirmwareErase %s\n", erasemsg.Kind, erasemsg.Data)

	log.Printf("Hash: %x\n", hash)
	deviceFirmwareUpload := &messages.FirmwareUpload{
		Payload: payload,
		Hash:    hash[:],
	}

	uploaddata, err := proto.Marshal(deviceFirmwareUpload)
	if err != nil {
		return err
	}
	chunks = makeTrezorMessage(uploaddata, messages.MessageType_MessageType_FirmwareUpload)

	uploadmsg, err := sendToDevice(dev, chunks)
	if err != nil {
		return err
	}
	log.Printf("Success %d! FirmwareUpload %s\n", uploadmsg.Kind, uploadmsg.Data)

	// Send ButtonAck
	chunks, err = MessageButtonAck()
	if err != nil {
		return err
	}
	return sendToDeviceNoAnswer(dev, chunks)
}

// DeviceSetMnemonic Configure the device with a mnemonic.
func DeviceSetMnemonic(deviceType DeviceType, mnemonic string) (wire.Message, error) {
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}
	defer dev.Close()

	// Send SetMnemonic

	skycoinSetMnemonic := &messages.SetMnemonic{
		Mnemonic: proto.String(mnemonic),
	}

	data, err := proto.Marshal(skycoinSetMnemonic)
	if err != nil {
		return wire.Message{}, err
	}

	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_SetMnemonic)

	msg, err := sendToDevice(dev, chunks)
	if err != nil {
		return wire.Message{}, err
	}

	if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
		msg, err = deviceButtonAck(dev)
		if err != nil {
			return wire.Message{}, err
		}
	}

	return msg, err
}

// DeviceGenerateMnemonic Ask the device to generate a mnemonic and configure itself with it.
func DeviceGenerateMnemonic(deviceType DeviceType, wordCount uint32, usePassphrase bool) (wire.Message, error) {
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}
	defer dev.Close()

	skycoinGenerateMnemonic := &messages.GenerateMnemonic{
		PassphraseProtection: proto.Bool(usePassphrase),
		WordCount:            proto.Uint32(wordCount),
	}

	data, err := proto.Marshal(skycoinGenerateMnemonic)
	if err != nil {
		return wire.Message{}, err
	}

	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_GenerateMnemonic)

	msg, err := sendToDevice(dev, chunks)
	if err != nil {
		return wire.Message{}, err
	}

	if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
		msg, err = deviceButtonAck(dev)
		if err != nil {
			return wire.Message{}, err
		}
	}

	return msg, err
}

func DecodeSuccessOrFailMsg(msg wire.Message) (string, error) {
	if msg.Kind == uint16(messages.MessageType_MessageType_Success) {
		return DecodeSuccessMsg(msg)
	}
	if msg.Kind == uint16(messages.MessageType_MessageType_Failure) {
		return DecodeFailMsg(msg)
	}

	return "", fmt.Errorf("Calling DecodeSuccessOrFailMsg on message kind %s", messages.MessageType(msg.Kind))
}

// DecodeSuccessMsg convert byte data into string containing the success message returned by the device
func DecodeSuccessMsg(msg wire.Message) (string, error) {
	if msg.Kind == uint16(messages.MessageType_MessageType_Success) {
		success := &messages.Success{}
		err := proto.Unmarshal(msg.Data, success)
		if err != nil {
			return "", err
		}
		return success.GetMessage(), nil
	}

	return "", fmt.Errorf("calling DecodeSuccessMsg with wrong message type: %s", messages.MessageType(msg.Kind))
}

// DecodeFailMsg convert byte data into string containing the failure returned by the device
func DecodeFailMsg(msg wire.Message) (string, error) {
	if msg.Kind == uint16(messages.MessageType_MessageType_Failure) {
		failure := &messages.Failure{}
		err := proto.Unmarshal(msg.Data, failure)
		if err != nil {
			return "", err
		}
		return failure.GetMessage(), nil
	}
	return "", fmt.Errorf("calling DecodeFailMsg with wrong message type: %s", messages.MessageType(msg.Kind))
}

// DecodeResponseSkycoinAddress convert byte data into list of addresses, meant to be used after DevicePinMatrixAck
func DecodeResponseSkycoinAddress(msg wire.Message) ([]string, error) {
	log.Printf("%x\n", msg.Data)

	if msg.Kind == uint16(messages.MessageType_MessageType_ResponseSkycoinAddress) {
		responseSkycoinAddress := &messages.ResponseSkycoinAddress{}
		err := proto.Unmarshal(msg.Data, responseSkycoinAddress)
		if err != nil {
			return []string{}, err
		}
		return responseSkycoinAddress.GetAddresses(), nil
	}

	return []string{}, fmt.Errorf("calling DecodeResponseSkycoinAddress with wrong message type: %s", messages.MessageType(msg.Kind))
}

// DecodeResponseTransactionSign convert byte data into list of signatures
func DecodeResponseTransactionSign(msg wire.Message) ([]string, error) {
	if msg.Kind == uint16(messages.MessageType_MessageType_ResponseTransactionSign) {
		responseSkycoinTransactionSign := &messages.ResponseTransactionSign{}
		err := proto.Unmarshal(msg.Data, responseSkycoinTransactionSign)
		if err != nil {
			return make([]string, 0), err
		}
		return responseSkycoinTransactionSign.GetSignatures(), nil
	}

	return []string{}, fmt.Errorf("calling DecodeResponseeSkycoinSignMessage with wrong message type: %s", messages.MessageType(msg.Kind))
}

// DecodeResponseSkycoinSignMessage convert byte data into signed message, meant to be used after DevicePinMatrixAck
func DecodeResponseSkycoinSignMessage(msg wire.Message) (string, error) {
	if msg.Kind == uint16(messages.MessageType_MessageType_ResponseSkycoinSignMessage) {
		responseSkycoinSignMessage := &messages.ResponseSkycoinSignMessage{}
		err := proto.Unmarshal(msg.Data, responseSkycoinSignMessage)
		if err != nil {
			return "", err
		}
		return responseSkycoinSignMessage.GetSignedMessage(), nil
	}
	return "", fmt.Errorf("calling DecodeResponseeSkycoinSignMessage with wrong message type: %s", messages.MessageType(msg.Kind))
}

// DeviceAddressGen Ask the device to generate an address
func DeviceAddressGen(deviceType DeviceType, addressN int, startIndex int, confirmAddress bool) (wire.Message, error) {
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}
	defer dev.Close()
	skycoinAddress := &messages.SkycoinAddress{
		AddressN:       proto.Uint32(uint32(addressN)),
		ConfirmAddress: proto.Bool(confirmAddress),
		StartIndex:     proto.Uint32(uint32(startIndex)),
	}
	data, err := proto.Marshal(skycoinAddress)
	if err != nil {
		return wire.Message{}, err
	}

	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_SkycoinAddress)

	return sendToDevice(dev, chunks)
}

// DeviceTransactionSign Ask the device to sign a transaction using the given information.
func DeviceTransactionSign(deviceType DeviceType, inputs []*messages.SkycoinTransactionInput, outputs []*messages.SkycoinTransactionOutput) (wire.Message, error) {
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}
	defer dev.Close()

	skycoinTransactionSignMessage := &messages.TransactionSign{
		NbIn:           proto.Uint32(uint32(len(inputs))),
		NbOut:          proto.Uint32(uint32(len(outputs))),
		TransactionIn:  inputs,
		TransactionOut: outputs,
	}
	log.Println(skycoinTransactionSignMessage)

	data, err := proto.Marshal(skycoinTransactionSignMessage)
	if err != nil {
		return wire.Message{}, err
	}

	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_TransactionSign)

	return sendToDevice(dev, chunks)
}

// DeviceSignMessage Ask the device to sign a message using the secret key at given index.
func DeviceSignMessage(deviceType DeviceType, addressN int, message string) (wire.Message, error) {
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}
	defer dev.Close()

	skycoinSignMessage := &messages.SkycoinSignMessage{
		AddressN: proto.Uint32(uint32(addressN)),
		Message:  proto.String(message),
	}

	data, err := proto.Marshal(skycoinSignMessage)
	if err != nil {
		return wire.Message{}, err
	}

	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_SkycoinSignMessage)

	return sendToDevice(dev, chunks)
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
func initialize(dev io.ReadWriteCloser) error {
	var chunks [][64]byte

	initialize := &messages.Initialize{}
	data, err := proto.Marshal(initialize)
	if err != nil {
		return err
	}

	chunks = makeTrezorMessage(data, messages.MessageType_MessageType_Initialize)
	_, err = sendToDevice(dev, chunks)

	return err
}

// DeviceApplySettings send ApplySettings request to the device
func DeviceApplySettings(deviceType DeviceType, usePassphrase bool, label string) (wire.Message, error) {
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}
	defer dev.Close()

	applySettings := &messages.ApplySettings{
		Label:         proto.String(label),
		Language:      proto.String(""),
		UsePassphrase: proto.Bool(usePassphrase),
	}
	log.Println(applySettings)
	data, err := proto.Marshal(applySettings)
	if err != nil {
		return wire.Message{}, err
	}

	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_ApplySettings)
	return sendToDevice(dev, chunks)
}

// DeviceGetFeatures send Features message to the device
func DeviceGetFeatures(deviceType DeviceType) (wire.Message, error) {
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}
	defer dev.Close()

	featureMsg := &messages.GetFeatures{}
	data, err := proto.Marshal(featureMsg)
	if err != nil {
		return wire.Message{}, err
	}

	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_GetFeatures)

	return sendToDevice(dev, chunks)
}

// BackupDevice ask the device to perform the seed backup
func BackupDevice(deviceType DeviceType) (wire.Message, error) {
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}
	defer dev.Close()
	var msg wire.Message
	var chunks [][64]byte
	err = initialize(dev)
	if err != nil {
		return wire.Message{}, err
	}

	backupDevice := &messages.BackupDevice{}
	data, err := proto.Marshal(backupDevice)
	if err != nil {
		return wire.Message{}, err
	}
	chunks = makeTrezorMessage(data, messages.MessageType_MessageType_BackupDevice)
	msg, err = sendToDevice(dev, chunks)
	if err != nil {
		return wire.Message{}, err
	}

	for msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
		msg, err = deviceButtonAck(dev)
		if err != nil {
			return wire.Message{}, err
		}
	}

	return msg, nil
}

// DeviceWordAck send a word to the device during device "recovery procedure"
func DeviceWordAck(deviceType DeviceType, word string) (wire.Message, error) {
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}

	defer dev.Close()
	chunks, err := MessageWordAck(word)
	if err != nil {
		return wire.Message{}, err
	}
	msg, err := sendToDevice(dev, chunks)
	if err != nil {
		return wire.Message{}, err
	}

	return msg, nil
}

// RecoveryDevice ask the device to perform the seed backup
func RecoveryDevice(deviceType DeviceType, wordCount uint32, usePassphrase, dryRun bool) (wire.Message, error) {
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}
	defer dev.Close()
	var msg wire.Message
	var chunks [][64]byte

	log.Printf("Using passphrase %t\n", usePassphrase)

	recoveryDevice := &messages.RecoveryDevice{
		WordCount:            proto.Uint32(wordCount),
		PassphraseProtection: proto.Bool(usePassphrase),
		DryRun:               proto.Bool(dryRun),
	}
	data, err := proto.Marshal(recoveryDevice)
	if err != nil {
		return wire.Message{}, err
	}
	chunks = makeTrezorMessage(data, messages.MessageType_MessageType_RecoveryDevice)
	msg, err = sendToDevice(dev, chunks)
	if err != nil {
		return msg, err
	}
	log.Printf("Recovery device %d! Answer is: %s\n", msg.Kind, msg.Data)

	if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
		msg, err = deviceButtonAck(dev)
		if err != nil {
			return wire.Message{}, err
		}
	}

	return msg, nil
}

// WipeDevice wipes out device configuration
func WipeDevice(deviceType DeviceType) (wire.Message, error) {
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}

	defer dev.Close()
	var chunks [][64]byte

	err = initialize(dev)
	if err != nil {
		return wire.Message{}, err
	}

	wipeDevice := &messages.WipeDevice{}
	data, err := proto.Marshal(wipeDevice)
	if err != nil {
		return wire.Message{}, err
	}

	var msg wire.Message
	chunks = makeTrezorMessage(data, messages.MessageType_MessageType_WipeDevice)
	msg, err = sendToDevice(dev, chunks)
	if err != nil {
		return wire.Message{}, err
	}
	log.Printf("Wipe device %d! Answer is: %x\n", msg.Kind, msg.Data)

	if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
		msg, err = deviceButtonAck(dev)
		if err != nil {
			return wire.Message{}, err
		}
	}

	if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
		err = initialize(dev)
		if err != nil {
			return wire.Message{}, err
		}
	}

	return msg, err
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
func DeviceChangePin(deviceType DeviceType) (wire.Message, error) {
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}
	defer dev.Close()

	changePin := &messages.ChangePin{}
	data, _ := proto.Marshal(changePin)
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_ChangePin)
	msg, err := sendToDevice(dev, chunks)
	if err != nil {
		return wire.Message{}, err
	}

	// Acknowledge that a button has been pressed
	if msg.Kind == uint16(messages.MessageType_MessageType_ButtonRequest) {
		msg, err = deviceButtonAck(dev)
		if err != nil {
			return msg, err
		}
	}
	return msg, nil
}

// DevicePinMatrixAck during PIN code setting use this message to send user input to device
func DevicePinMatrixAck(deviceType DeviceType, p string) (wire.Message, error) {
	time.Sleep(1 * time.Second)
	dev, err := getDevice(deviceType)
	if err != nil {
		return wire.Message{}, err
	}
	defer dev.Close()

	log.Printf("Setting pin: %s\n", p)
	pinAck := &messages.PinMatrixAck{
		Pin: proto.String(p),
	}
	data, err := proto.Marshal(pinAck)
	if err != nil {
		return wire.Message{}, err
	}

	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_PinMatrixAck)
	return sendToDevice(dev, chunks)
}
