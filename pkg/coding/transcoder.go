package coding

import (
	"encoding/json"
	"io"

	"github.com/linkedin/goavro"
	"github.com/pkg/errors"
)

type SchemaLookup func(int64) (string, error)

func NewJsonTranscoder(lookup SchemaLookup, out io.Writer) *JsonTranscoder {
	var coder JsonTranscoder
	coder.encoder = json.NewEncoder(out)
	coder.lookup = lookup
	return &coder
}

type JsonTranscoder struct {
	cachedCodecs map[int64]*goavro.Codec
	encoder      *json.Encoder
	lookup       SchemaLookup
}

func (coder *JsonTranscoder) Transcode(schemaID int64, data []byte) error {
	codec, ok := coder.cachedCodecs[schemaID]
	if !ok {
		schema, err := coder.lookup(schemaID)
		if err != nil {
			return err
		}

		codec, err = goavro.NewCodec(schema)
		if err != nil {
			return errors.Wrapf(err, "failed to parse Avro schema with ID %d", schemaID)
		}
		if coder.cachedCodecs == nil {
			coder.cachedCodecs = make(map[int64]*goavro.Codec, 16)
		}

		coder.cachedCodecs[schemaID] = codec
	}

	decoded, _, err := codec.NativeFromBinary(data)
	if err != nil {
		return errors.Wrapf(err, "failed to decode message using schema with ID %d", schemaID)
	}

	err = coder.encoder.Encode(decoded)
	if err != nil {
		return errors.Wrap(err, "failed to encode message as JSON")
	}

	return nil
}
