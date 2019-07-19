package go_parquet

import (
	"encoding/binary"
	"io"
)

type int64PlainDecoder struct {
	r io.Reader
}

func (i *int64PlainDecoder) init(r io.Reader) error {
	i.r = r

	return nil
}

func (i *int64PlainDecoder) decodeValues(dst []interface{}) error {
	d := make([]int64, len(dst))
	if err := binary.Read(i.r, binary.LittleEndian, d); err != nil {
		return err
	}
	for i := range d {
		dst[i] = d[i]
	}
	return nil
}

type int64PlainEncoder struct {
	w io.Writer
}

func (i *int64PlainEncoder) Close() error {
	return nil
}

func (i *int64PlainEncoder) init(w io.Writer) error {
	i.w = w

	return nil
}

func (i *int64PlainEncoder) encodeValues(values []interface{}) error {
	d := make([]int64, len(values))
	for i := range values {
		d[i] = values[i].(int64)
	}

	return binary.Write(i.w, binary.LittleEndian, d)
}

type int64DeltaBPDecoder struct {
	deltaBitPackDecoder64
}

func (d *int64DeltaBPDecoder) decodeValues(dst []interface{}) error {
	for i := range dst {
		u, err := d.next()
		if err != nil {
			return err
		}
		dst[i] = u
	}

	return nil
}

type int64DeltaBPEncoder struct {
	deltaBitPackEncoder64
}

func (d *int64DeltaBPEncoder) encodeValues(values []interface{}) error {
	for i := range values {
		if err := d.addInt64(values[i].(int64)); err != nil {
			return err
		}
	}

	return nil
}