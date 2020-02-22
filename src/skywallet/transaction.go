package skywallet

import (
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"

	messages "github.com/SkycoinProject/hardware-wallet-protob/go"

	"github.com/SkycoinProject/hardware-wallet-go/src/skywallet/wire"
)

//go:generate mockery -name TransactionSigner -case underscore -inpkg -testonly

// TransactionSigner represents a data (inputs, outputs) needed to sign transaction
type TransactionSigner interface {
	Sign() ([]string, error)
}

// SkycoinTransactionSigner represents signing Skycoin transaction process
// @used_in TransactionSign
type SkycoinTransactionSigner struct {
	Device     *Device
	Inputs     []*messages.TxAck_TransactionType_TxInputType
	Outputs    []*messages.TxAck_TransactionType_TxOutputType
	Version    int
	LockTime   int
	signatures []string
	state      int
}

// Sign method signs the Skycoin Transaction
func (s *SkycoinTransactionSigner) Sign() ([]string, error) {
	msg, err := s.initSigningProcess()

	if err != nil {
		return nil, err
	}

	index := 0
	s.state = 0

	for {
		switch msg.Kind {
		case uint16(messages.MessageType_MessageType_TxRequest):
			txRequest := &messages.TxRequest{}
			err = proto.Unmarshal(msg.Data, txRequest)
			if err != nil {
				return nil, err
			}
			switch *txRequest.RequestType {
			case messages.TxRequest_TXINPUT:
				if s.state == 0 { // Sending Inputs for InnerHash
					if len(s.Inputs)-index > 8 {
						msg, err = s.sendInputs(index, 8)
						if err != nil {
							return nil, err
						}
						index += 8
					} else {
						msg, err = s.sendInputs(index, len(s.Inputs)-index)
						if err != nil {
							return nil, err
						}
						s.state++
						index = 0
					}
				} else if s.state == 2 { // Sending Inputs for Signatures
					err = s.addSignatures(&msg)
					if err != nil {
						return nil, err
					}
					if len(s.Inputs)-index > 8 {
						msg, err = s.sendInputs(index, 8)
						if err != nil {
							return nil, err
						}
					} else {
						msg, err = s.sendInputs(index, len(s.Inputs)-index)
						if err != nil {
							return nil, err
						}
						s.state++
						index = 0
					}
					index += 8
				} else {
					return nil, fmt.Errorf("protocol error: unexpected TxRequest type")
				}
			case messages.TxRequest_TXOUTPUT:
				if s.state == 1 { // Sending Outputs for InnerHash
					if len(s.Outputs)-index > 8 {
						msg, err = s.sendOutputs(index, 8)
						if err != nil {
							return nil, err
						}
						index += 8
					} else {
						msg, err = s.sendOutputs(index, len(s.Outputs)-index)
						if err != nil {
							return nil, err
						}
						s.state++
						index = 0
					}
				} else {
					return nil, fmt.Errorf("protocol error: unexpected TxRequest type")
				}
			case messages.TxRequest_TXFINISHED:
				if s.state == 3 {
					err = s.addSignatures(&msg)
					if err != nil {
						return nil, err
					}
					return s.signatures, nil
				}
				return nil, fmt.Errorf("protocol error: unexpected TXFINISHED message")
			}
		case uint16(messages.MessageType_MessageType_Failure):
			failMsg, err := DecodeFailMsg(msg)
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("failed with message: %s", failMsg)
		case uint16(messages.MessageType_MessageType_ButtonRequest):
			msg, err = s.Device.ButtonAck()
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unexpected response message type from hardware wallet")
		}
	}
}

func (s *SkycoinTransactionSigner) initSigningProcess() (wire.Message, error) {
	return s.Device.SignTx(len(s.Outputs), len(s.Inputs), "Skycoin", s.Version, s.LockTime, "dkdji9e2oidhash")
}

func (s *SkycoinTransactionSigner) sendInputs(startIndex, count int) (wire.Message, error) {
	if startIndex+count > len(s.Inputs) {
		return wire.Message{}, fmt.Errorf("invalid index or count")
	}

	txInputs := s.Inputs[startIndex : startIndex+count]
	if len(txInputs) != 0 {
		return s.Device.TxAck(txInputs, nil, s.Version, s.LockTime)
	}
	return wire.Message{}, errors.New("empty inputs")
}

func (s *SkycoinTransactionSigner) sendOutputs(startIndex, count int) (wire.Message, error) {
	if startIndex+count > len(s.Outputs) {
		return wire.Message{}, fmt.Errorf("invalid index or count")
	}

	txOutputs := s.Outputs[startIndex : startIndex+count]
	if len(txOutputs) != 0 {
		return s.Device.TxAck(nil, txOutputs, s.Version, s.LockTime)
	}
	return wire.Message{}, errors.New("empty inputs")
}

func (s *SkycoinTransactionSigner) addSignatures(msg *wire.Message) error {
	txRequest := &messages.TxRequest{}
	err := proto.Unmarshal(msg.Data, txRequest)
	if err != nil {
		return err
	}
	for _, sign := range txRequest.SignResult {
		s.signatures = append(s.signatures, sign.GetSignature())
	}
	return nil
}
