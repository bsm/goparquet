package go_parquet

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/pkg/errors"
)

type byteReader struct {
	io.Reader
}

func (br *byteReader) ReadByte() (byte, error) {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(br.Reader, buf); err != nil {
		return 0, err
	}

	return buf[0], nil
}

type offsetReader struct {
	inner  io.ReadSeeker
	offset int64
	count  int64
}

func (o *offsetReader) Read(p []byte) (int, error) {
	n, err := o.inner.Read(p)
	o.offset += int64(n)
	o.count += int64(n)
	return n, err
}

func (o *offsetReader) Seek(offset int64, whence int) (int64, error) {
	i, err := o.inner.Seek(offset, whence)
	if err == nil {
		o.count += i - o.offset
		o.offset = i
	}

	return i, err
}

func (o *offsetReader) Count() int64 {
	return o.count
}

func decodeRLEValue(bytes []byte) int32 {
	switch len(bytes) {
	case 1:
		return int32(bytes[0])
	case 2:
		return int32(bytes[0]) + int32(bytes[1])<<8
	case 3:
		return int32(bytes[0]) + int32(bytes[1])<<8 + int32(bytes[2])<<16
	case 4:
		return int32(bytes[0]) + int32(bytes[1])<<8 + int32(bytes[2])<<16 + int32(bytes[3])<<24
	default:
		panic("invalid argument")
	}
}

func encodeRLEValue(in int32, size int) []byte {
	switch size {
	case 1:
		return []byte{byte(in & 255)}
	case 2:
		return []byte{
			byte(in & 255),
			byte((in >> 8) & 255),
		}
	case 3:
		return []byte{
			byte(in & 255),
			byte((in >> 8) & 255),
			byte((in >> 16) & 255),
		}
	case 4:
		return []byte{
			byte(in & 255),
			byte((in >> 8) & 255),
			byte((in >> 16) & 255),
			byte((in >> 24) & 255),
		}
	default:
		panic("invalid argument")
	}
}

func writeFull(w io.Writer, buf []byte) error {
	cnt, err := w.Write(buf)
	if err != nil {
		return err
	}

	if cnt != len(buf) {
		return errors.Errorf("need to write %d byte wrote %d", cnt, len(buf))
	}

	return nil
}

type thriftReader interface {
	Read(thrift.TProtocol) error
}

func readThrift(tr thriftReader, r io.Reader) error {
	// Make sure we are not using any kind of buffered reader here. bufio.Reader "can" reads more data ahead of time,
	// which is a problem on this library
	transport := &thrift.StreamTransport{Reader: r}
	proto := thrift.NewTCompactProtocol(transport)
	return tr.Read(proto)
}

func decodeUint16(d decoder, data []uint16) error {
	for i := range data {
		u, err := d.next()
		if err != nil {
			return err
		}
		data[i] = uint16(u)
	}

	return nil
}

func decodeInt32(d decoder, data []int32) error {
	for i := range data {
		u, err := d.next()
		if err != nil {
			return err
		}
		data[i] = u
	}

	return nil
}

func readUVariant32(r io.Reader) (int32, error) {
	b, ok := r.(io.ByteReader)
	if !ok {
		b = &byteReader{Reader: r}
	}

	i, err := binary.ReadUvarint(b)
	if err != nil {
		return 0, err
	}

	if i > math.MaxInt32 {
		return 0, errors.New("int32 out of range")
	}

	return int32(i), nil
}

func readVariant32(r io.Reader) (int32, error) {
	b, ok := r.(io.ByteReader)
	if !ok {
		b = &byteReader{Reader: r}
	}

	i, err := binary.ReadVarint(b)
	if err != nil {
		return 0, err
	}

	if i > math.MaxInt32 || i < math.MinInt32 {
		return 0, errors.New("int32 out of range")
	}

	return int32(i), nil
}

func readVariant64(r io.Reader) (int64, error) {
	b, ok := r.(io.ByteReader)
	if !ok {
		b = &byteReader{Reader: r}
	}

	return binary.ReadVarint(b)
}

type constDecoder int32

func (cd constDecoder) initSize(io.Reader) error {
	return nil
}

func (cd constDecoder) init(io.Reader) error {
	return nil
}

func (cd constDecoder) next() (int32, error) {
	return int32(cd), nil
}

type levelDecoderWrapper struct {
	decoder
	max uint16
}

func (l *levelDecoderWrapper) maxLevel() uint16 {
	return l.max
}
