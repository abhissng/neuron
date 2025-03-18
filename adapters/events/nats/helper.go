package nats

import (
	"github.com/abhissng/neuron/utils/codec"
	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/nats-io/nats.go"
)

// EncodedNatsMsg returns an encoded message from a nats.Msg.
func EncodedNatsMsg(msg *nats.Msg) string {
	message := map[string]any{}
	message["subject"] = msg.Subject
	message["reply"] = msg.Reply
	message["header"] = msg.Header
	message["data"] = string(msg.Data)

	byt, err := codec.Encode(message, codec.JSON)
	if err != nil {
		helpers.Println(constant.ERROR, "failed to encode nats msg: "+err.Error())
		return ""
	}

	return string(byt)
}
