package codec

import "github.com/abhissng/neuron/utils/types"

// for encoding and decoding
const (
	// Text-based formats
	JSON types.CodecType = "json"
	XML  types.CodecType = "xml"
	YAML types.CodecType = "yaml"
	TOML types.CodecType = "toml"
	INI  types.CodecType = "ini"

	// Binary formats
	Gob         types.CodecType = "gob"
	ProtoBuf    types.CodecType = "protobuf"
	MessagePack types.CodecType = "msgpack"
	CBOR        types.CodecType = "cbor"
	Avro        types.CodecType = "avro"

	// Encoding schemes
	Base64 types.CodecType = "base64"
	Base32 types.CodecType = "base32"
	Base16 types.CodecType = "base16"
	Hex    types.CodecType = "hex"
	URL    types.CodecType = "url"

	// Compression formats
	Gzip   types.CodecType = "gzip"
	Zlib   types.CodecType = "zlib"
	Brotli types.CodecType = "brotli"
	LZ4    types.CodecType = "lz4"
	Snappy types.CodecType = "snappy"
)
