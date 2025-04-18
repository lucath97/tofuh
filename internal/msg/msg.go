package msg

import (
	"errors"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"lucathurm.dev/tofuh/internal/config"
)

const (
	maxPos    byte = 63
	unsetMask byte = 0b01111111
	setMask   byte = 0b10000000
)

var (
	ErrClientNotConnected = errors.New("mqtt client not connected")
	ErrMsgMalfunction     = errors.New("internal messaging malfunction")
	ErrPosOutOfRange      = errors.New("position out of range 0 - 63")
)

// binary message containing a 64 bit state
type StateMsg [8]byte

// binary message format containing info for setting or unsetting a bit in a 64 bit mask
// set (2⁷): new value for the bit (0-1)
// pos (2⁰-2⁶): position of the bit to set (0-63)
type SetBitMsg [1]byte

func (m SetBitMsg) Unmarshal() (byte, bool) {
	pos := m[0] & unsetMask
	set := m[0]>>7 != 0

	return pos, set
}

func NewSetBitMsg(pos byte, set bool) (SetBitMsg, error) {
	var msg SetBitMsg
	if pos > maxPos {
		return msg, ErrPosOutOfRange
	}

	msg[0] |= pos

	if set {
		msg[0] |= setMask
	} else {
		msg[0] &= unsetMask
	}

	return msg, nil
}

func NewClient(cfg *config.Config) *mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MsgAddress)
	return nil
}

func Publish(client *mqtt.Client, topic string, qos byte, retained bool, msg []byte) error {
	if !(*client).IsConnected() {
		return ErrClientNotConnected
	}

	token := (*client).Publish(topic, qos, retained, msg)
	if token.Error() != nil {
		return ErrMsgMalfunction
	}

	return nil
}

func Subscribe(client *mqtt.Client, topic string, qos byte, callback func(c mqtt.Client, m mqtt.Message)) error {
	if !(*client).IsConnected() {
		return ErrClientNotConnected
	}

	token := (*client).Subscribe(topic, qos, callback)
	if token.Error() != nil {
		return ErrMsgMalfunction
	}

	return nil
}
