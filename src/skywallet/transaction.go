package skywallet

import (
	"errors"
	"fmt"
	"github.com/SkycoinProject/hardware-wallet-go/src/skywallet/wire"
	messages "github.com/SkycoinProject/hardware-wallet-protob/go"
	"github.com/gogo/protobuf/proto"
)

//go:generate mockery -name TransactionSigner -case underscore -inpkg -testonly

// TransactionData represents a data (inputs, outputs) needed to sign transaction
type TransactionSigner interface {
	InitSigningProcess() (wire.Message, error)
	SendInputs(startIndex, count int) (wire.Message, error)
	SendOutputs(startIndex, count int) (wire.Message, error)
	GetSignatures() []string
}

// *
// Structure representing signing Skycoin transaction process
// @used_in TransactionSign
type SkycoinTransactionSigner struct {
	Device   *Device
	Inputs   []*messages.TxAck_TransactionType_TxInputType
	Outputs  []*messages.TxAck_TransactionType_TxOutputType
	Signatures []string
	Version  int
	LockTime int
	State    int
}

func (s *SkycoinTransactionSigner) InitSigningProcess() (wire.Message, error) {
	return s.Device.SignTx(len(s.Outputs), len(s.Inputs), "Skycoin", s.Version, s.LockTime, "dkdji9e2oidhash")
}

func (s *SkycoinTransactionSigner) SendInputs(startIndex, count int) (wire.Message, error) {
	if startIndex+count > len(s.Inputs) {
		return wire.Message{}, fmt.Errorf("invalid index or count")
	}

	txInputs := s.Inputs[startIndex : startIndex+count]
	if len(txInputs) != 0 {
		return s.Device.TxAck(txInputs, nil, s.Version, s.LockTime)
	}
	return wire.Message{}, errors.New("empty inputs")
}

func (s *SkycoinTransactionSigner) SendOutputs(startIndex, count int) (wire.Message, error) {
	if startIndex+count > len(s.Outputs) {
		return wire.Message{}, fmt.Errorf("invalid index or count")
	}

	txOutputs := s.Outputs[startIndex : startIndex+count]
	if len(txOutputs) != 0 {
		return s.Device.TxAck(nil, txOutputs, s.Version, s.LockTime)
	}
	return wire.Message{}, errors.New("empty inputs")
}

func (s *SkycoinTransactionSigner) AddSignatures(msg *wire.Message) error {
	txRequest := &messages.TxRequest{}
	err := proto.Unmarshal(msg.Data, txRequest)
	if err != nil {
		return err
	}
	for _, sign := range txRequest.SignResult {
		s.Signatures = append(s.Signatures, sign.GetSignature())
	}
	return nil
}
// *
// Structure representing signing Bitcoin transaction process
// @used_in TransactionSign
type BitcoinTransactionSigner struct {
	device  *Device
	inputs  []messages.BitcoinTransactionInput
	outputs []messages.BitcoinTransactionOutput
	state   int
}
