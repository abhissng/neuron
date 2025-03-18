package http

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"

	"github.com/abhissng/neuron/utils/types"
	"github.com/vmihailenco/msgpack/v5"
	"gopkg.in/yaml.v3"
)

// Decoder function type
type DecoderFunc func(io.Reader, any) error

// Decoder registry to map content types to decoding functions
var decoders = map[types.ContentType]DecoderFunc{
	ContentTypeJSON:      decodeJSON,
	ContentTypeXML:       decodeXML,
	ContentTypeYAML:      decodeYAML,
	ContentTypeMsgPack:   decodeMsgPack,
	ContentTypePlainText: decodeText,
}

// Decoder functions for different formats

// Decodes JSON
func decodeJSON(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

// Decodes XML
func decodeXML(r io.Reader, v any) error {
	return xml.NewDecoder(r).Decode(v)
}

// Decodes YAML
func decodeYAML(r io.Reader, v any) error {
	return yaml.NewDecoder(r).Decode(v)
}

// Decodes MsgPack
func decodeMsgPack(r io.Reader, v any) error {
	return msgpack.NewDecoder(r).Decode(v)
}

// Decodes text
func decodeText(r io.Reader, v any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	str, ok := v.(*string)
	if !ok {
		return errors.New("expected *string for text decoding")
	}
	*str = string(b)
	return nil
}
