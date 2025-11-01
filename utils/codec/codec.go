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
	"fmt"
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
		var decoded []byte
		decoded, err = base64.StdEncoding.DecodeString(toString(data))
		if err == nil {
			err = decodeBytes(decoded, &result)
		}

	case Hex:
		var decoded []byte
		decoded, err = hex.DecodeString(toString(data))
		if err == nil {
			err = decodeBytes(decoded, &result)
		}

	case Gzip:
		buf := bytes.NewBuffer(data)
		var gz *gzip.Reader
		gz, err = gzip.NewReader(buf)
		if err == nil {
			defer func() {
				_ = gz.Close()
			}()
			var decodedData []byte
			decodedData, err = io.ReadAll(gz)
			if err == nil {
				err = decodeBytes(decodedData, &result)
			}
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

// decodeBytes handles assigning decoded byte data to result using toString.
func decodeBytes[T any](decoded []byte, result *T) error {
	// Use toString() to make it compatible with both string and []byte targets
	switch any(*result).(type) {
	case string:
		*result = any(toString(decoded)).(T)
	case []byte:
		*result = any([]byte(toString(decoded))).(T)
	default:
		// For structs, maps, slices, etc. try JSON unmarshal
		if err := json.Unmarshal(decoded, result); err != nil {
			var zero T
			*result = zero
			return fmt.Errorf("decode: unable to map decoded bytes to target type: %w", err)
		}
	}
	return nil
}
