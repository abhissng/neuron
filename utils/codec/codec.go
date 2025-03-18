package codec

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/helpers"
	"github.com/abhissng/neuron/utils/types"
	"gopkg.in/yaml.v3"
)

// Encode serializes data based on the codec type.
func Encode[T any](data T, codecType types.CodecType) ([]byte, error) {
	var buf bytes.Buffer
	var err error

	switch codecType {
	case JSON:
		err = json.NewEncoder(&buf).Encode(data)
	case XML:
		err = xml.NewEncoder(&buf).Encode(data)
	case YAML:
		err = yaml.NewEncoder(&buf).Encode(data)
	case Gob:
		enc := gob.NewEncoder(&buf)
		err = enc.Encode(data)
	case Base64:
		encoded := base64.StdEncoding.EncodeToString([]byte(toString(data)))
		return []byte(encoded), nil
	case Hex:
		encoded := hex.EncodeToString([]byte(toString(data)))
		return []byte(encoded), nil
	case Gzip:
		gz := gzip.NewWriter(&buf)
		_, err = gz.Write([]byte(toString(data)))

		defer func() {
			if err := gz.Close(); err != nil {
				helpers.Println(constant.ERROR, "Error closing gzip reader: ", err)
			}
		}()

	default:
		return nil, errors.New("unsupported encoding format")
	}

	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode deserializes data based on the codec type.
func Decode[T any](data []byte, codecType types.CodecType) (T, error) {
	var result T
	var err error

	switch codecType {
	case JSON:
		err = json.Unmarshal(data, &result)
	case XML:
		err = xml.Unmarshal(data, &result)
	case YAML:
		err = yaml.Unmarshal(data, &result)
	case Gob:
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		err = dec.Decode(&result)
	case Base64:
		decoded, err := base64.StdEncoding.DecodeString(string(data))
		if err == nil {
			result = any(decoded).(T)
		}
	case Hex:
		decoded, err := hex.DecodeString(string(data))
		if err == nil {
			result = any(decoded).(T)
		}
	case Gzip:
		buf := bytes.NewBuffer(data)
		gz, err := gzip.NewReader(buf)
		if err == nil {
			decodedData, _ := io.ReadAll(gz)
			result = any(decodedData).(T)
		}
	default:
		err = errors.New("unsupported decoding format")
	}

	return result, err
}

// toString converts any type to a string representation.
func toString[T any](data T) string {
	switch v := any(data).(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		bytes, _ := Encode(v, JSON)
		return string(bytes)
	}
}
