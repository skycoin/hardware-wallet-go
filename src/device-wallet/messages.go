package devicewallet

import (
	"crypto/rand"
	"fmt"

	"github.com/gogo/protobuf/proto"
	messages "github.com/skycoin/hardware-wallet-go/src/device-wallet/messages/go"
	"github.com/skycoin/hardware-wallet-go/src/device-wallet/wire"
)

// MessageCancel prepare Cancel request
func MessageCancel() ([][64]byte, error) {
	msg := &messages.Cancel{}
	data, err := proto.Marshal(msg)
	if err != nil {
		return [][64]byte{}, err
	}
	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_Cancel)
	return chunks, nil
}

// MessageButtonAck send this message (before user action) when the device expects the user to push a button
func MessageButtonAck() ([][64]byte, error) {
	buttonAck := &messages.ButtonAck{}
	data, err := proto.Marshal(buttonAck)
	if err != nil {
		return [][64]byte{}, err
	}
	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_ButtonAck)
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
	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_PassphraseAck)
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
	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_WordAck)
	return chunks, nil
}

// MessageCheckMessageSignature prepare CheckMessageSignature request
func MessageCheckMessageSignature(message, signature, address string) ([][64]byte, error) {
	msg := &messages.SkycoinCheckMessageSignature{
		Address:   proto.String(address),
		Message:   proto.String(message),
		Signature: proto.String(signature),
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return [][64]byte{}, err
	}
	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_Cancel)
	return chunks, nil
}

// MessageAddressGen prepare MessageAddressGen request
func MessageAddressGen(addressN, startIndex int, confirmAddress bool) ([][64]byte, error) {
	skycoinAddress := &messages.SkycoinAddress{
		AddressN:       proto.Uint32(uint32(addressN)),
		ConfirmAddress: proto.Bool(confirmAddress),
		StartIndex:     proto.Uint32(uint32(startIndex)),
	}

	data, err := proto.Marshal(skycoinAddress)
	if err != nil {
		return [][64]byte{}, err
	}

	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_SkycoinAddress)
	return chunks, nil
}

// MessageApplySettings prepare MessageApplySettings request
func MessageApplySettings(usePassphrase bool, label string) ([][64]byte, error) {
	applySettings := &messages.ApplySettings{
		Label:         proto.String(label),
		Language:      proto.String(""),
		UsePassphrase: proto.Bool(usePassphrase),
	}
	log.Println(applySettings)
	data, err := proto.Marshal(applySettings)
	if err != nil {
		return [][64]byte{}, err
	}

	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_ApplySettings)
	return chunks, nil
}

// MessageBackup prepare MessageBackup request
func MessageBackup() ([][64]byte, error) {
	backupDevice := &messages.BackupDevice{}
	data, err := proto.Marshal(backupDevice)
	if err != nil {
		return [][64]byte{}, err
	}
	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_BackupDevice)
	return chunks, nil
}

// MessageChangePin prepare MessageChangePin request
func MessageChangePin() ([][64]byte, error) {
	changePin := &messages.ChangePin{}
	data, err := proto.Marshal(changePin)
	if err != nil {
		return [][64]byte{}, err
	}
	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_ChangePin)
	return chunks, nil
}

// MessageConnected prepare MessageConnected request
func MessageConnected() ([][64]byte, error) {
	msgRaw := &messages.Ping{}
	data, err := proto.Marshal(msgRaw)
	if err != nil {
		return [][64]byte{}, err
	}
	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_Ping)
	return chunks, nil
}

// MessageFirmwareErase prepare MessageFirmwareErase request
func MessageFirmwareErase(payload []byte) ([][64]byte, error) {
	deviceFirmwareErase := &messages.FirmwareErase{
		Length: proto.Uint32(uint32(len(payload))),
	}

	erasedata, err := proto.Marshal(deviceFirmwareErase)
	if err != nil {
		return [][64]byte{}, err
	}

	chunks := makeSkyWalletMessage(erasedata, messages.MessageType_MessageType_FirmwareErase)
	return chunks, nil
}

// MessageFirmwareUpload prepare MessageFirmwareUpload request
func MessageFirmwareUpload(payload []byte, hash [32]byte) ([][64]byte, error) {
	deviceFirmwareUpload := &messages.FirmwareUpload{
		Payload: payload,
		Hash:    hash[:],
	}

	uploaddata, err := proto.Marshal(deviceFirmwareUpload)
	if err != nil {
		return [][64]byte{}, err
	}

	chunks := makeSkyWalletMessage(uploaddata, messages.MessageType_MessageType_FirmwareUpload)
	return chunks, nil
}

// MessageGetFeatures prepare MessageGetFeatures request
func MessageGetFeatures() ([][64]byte, error) {
	featureMsg := &messages.GetFeatures{}
	data, err := proto.Marshal(featureMsg)
	if err != nil {
		return [][64]byte{}, err
	}

	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_GetFeatures)
	return chunks, nil
}

// MessageGenerateMnemonic prepare MessageGenerateMnemonic request
func MessageGenerateMnemonic(wordCount uint32, usePassphrase bool) ([][64]byte, error) {
	skycoinGenerateMnemonic := &messages.GenerateMnemonic{
		PassphraseProtection: proto.Bool(usePassphrase),
		WordCount:            proto.Uint32(wordCount),
	}

	data, err := proto.Marshal(skycoinGenerateMnemonic)
	if err != nil {
		return [][64]byte{}, err
	}

	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_GenerateMnemonic)
	return chunks, nil
}

// MessageRecovery prepare MessageRecovery request
func MessageRecovery(wordCount uint32, usePassphrase, dryRun bool) ([][64]byte, error) {
	recoveryDevice := &messages.RecoveryDevice{
		WordCount:            proto.Uint32(wordCount),
		PassphraseProtection: proto.Bool(usePassphrase),
		DryRun:               proto.Bool(dryRun),
	}
	data, err := proto.Marshal(recoveryDevice)
	if err != nil {
		return [][64]byte{}, err
	}

	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_RecoveryDevice)

	return chunks, nil
}

// MessageSetMnemonic prepare MessageSetMnemonic request
func MessageSetMnemonic(mnemonic string) ([][64]byte, error) {
	skycoinSetMnemonic := &messages.SetMnemonic{
		Mnemonic: proto.String(mnemonic),
	}

	data, err := proto.Marshal(skycoinSetMnemonic)
	if err != nil {
		return [][64]byte{}, err
	}

	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_SetMnemonic)
	return chunks, nil
}

// MessageSignMessage prepare MessageSignMessage request
func MessageSignMessage(addressN int, message string) ([][64]byte, error) {
	skycoinSignMessage := &messages.SkycoinSignMessage{
		AddressN: proto.Uint32(uint32(addressN)),
		Message:  proto.String(message),
	}

	data, err := proto.Marshal(skycoinSignMessage)
	if err != nil {
		return [][64]byte{}, err
	}

	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_SkycoinSignMessage)
	return chunks, nil
}

// MessageTransactionSign prepare MessageTransactionSign request
func MessageTransactionSign(inputs []*messages.SkycoinTransactionInput, outputs []*messages.SkycoinTransactionOutput) ([][64]byte, error) {
	skycoinTransactionSignMessage := &messages.TransactionSign{
		NbIn:           proto.Uint32(uint32(len(inputs))),
		NbOut:          proto.Uint32(uint32(len(outputs))),
		TransactionIn:  inputs,
		TransactionOut: outputs,
	}
	log.Println(skycoinTransactionSignMessage)

	data, err := proto.Marshal(skycoinTransactionSignMessage)
	if err != nil {
		return [][64]byte{}, err
	}

	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_TransactionSign)
	return chunks, nil
}

// MessageWipe prepare MessageWipe request
func MessageWipe() ([][64]byte, error) {
	wipeDevice := &messages.WipeDevice{}
	data, err := proto.Marshal(wipeDevice)
	if err != nil {
		return [][64]byte{}, err
	}

	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_WipeDevice)
	return chunks, nil
}

// MessagePinMatrixAck prepare MessagePinMatrixAck request
func MessagePinMatrixAck(p string) ([][64]byte, error) {
	pinAck := &messages.PinMatrixAck{
		Pin: proto.String(p),
	}
	data, err := proto.Marshal(pinAck)
	if err != nil {
		return [][64]byte{}, err
	}

	chunks := makeSkyWalletMessage(data, messages.MessageType_MessageType_PinMatrixAck)
	return chunks, nil
}

func MessageEntropyAck(bufferSize int) ([][64]byte, error) {
	buffer := make([]byte, bufferSize)
	if len, err := rand.Read(buffer); err != nil || len != bufferSize {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("required %d bytes but got %d", bufferSize, len)
	}
	entropyAck := &messages.EntropyAck{
		Entropy: buffer,
	}
	data, err := proto.Marshal(entropyAck)
	if err != nil {
		return nil, err
	}
	chunks := makeTrezorMessage(data, messages.MessageType_MessageType_EntropyAck)
	return chunks, nil
}

func DecodeSuccessOrFailMsg(msg wire.Message) (string, error) {
	if msg.Kind == uint16(messages.MessageType_MessageType_Success) {
		return DecodeSuccessMsg(msg)
	}
	if msg.Kind == uint16(messages.MessageType_MessageType_Failure) {
		return DecodeFailMsg(msg)
	}

	return "", fmt.Errorf("calling DecodeSuccessOrFailMsg on message kind %s", messages.MessageType(msg.Kind))
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
