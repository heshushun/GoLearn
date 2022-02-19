package proto

import (
	"GoLearn/goredis/core/bufio2"
	"bytes"
	"io"
	"log"
	"strconv"

	"errors"
)

var (
	ErrBadArrayLen        = errors.New("bad array len")
	ErrBadArrayLenTooLong = errors.New("bad array len, too long")

	ErrBadBulkBytesLen        = errors.New("bad bulk bytes len")
	ErrBadBulkBytesLenTooLong = errors.New("bad bulk bytes len, too long")

	ErrBadMultiBulkLen     = errors.New("bad multi-bulk len")
	ErrBadMultiBulkContent = errors.New("bad multi-bulk content, should be bulkbytes")
)

const (
	// MaxBulkBytesLen 最大长度
	MaxBulkBytesLen = 1024 * 1024 * 512
	// MaxArrayLen 最大长度
	MaxArrayLen = 1024 * 1024
)

type RespType byte

const (
	TypeString    = '+'
	TypeError     = '-'
	TypeInt       = ':'
	TypeBulkBytes = '$'
	TypeArray     = '*'
)

// Btoi64 byte to int64
func Btoi64(b []byte) (int64, error) {
	if len(b) != 0 && len(b) < 10 {
		var neg, i = false, 0
		switch b[0] {
		case '-':
			neg = true
			fallthrough
		case '+':
			i++
		}
		if len(b) != i {
			var n int64
			for ; i < len(b) && b[i] >= '0' && b[i] <= '9'; i++ {
				n = int64(b[i]-'0') + n*10
			}
			if len(b) == i {
				if neg {
					n = -n
				}
				return n, nil
			}
		}
	}

	if n, err := strconv.ParseInt(string(b), 10, 64); err != nil {
		return 0, errorsTrace(err)
	} else {
		return n, nil
	}
}

/*
*
* Encoder 编码器
*
**/
type Encoder struct {
	bw *bufio2.Writer

	Err error
}

func NewEncoder(w io.Writer) *Encoder {
	return NewEncoderBuffer(bufio2.NewWriterSize(w, 8192))
}

func NewEncoderSize(w io.Writer, size int) *Encoder {
	return NewEncoderBuffer(bufio2.NewWriterSize(w, size))
}

func NewEncoderBuffer(bw *bufio2.Writer) *Encoder {
	return &Encoder{bw: bw}
}

func (e *Encoder) Encode(r *Resp, flush bool) error {
	if e.Err != nil {
		return errorsTrace(e.Err)
	}
	if err := e.encodeResp(r); err != nil {
		e.Err = err
	} else if flush {
		e.Err = errorsTrace(e.bw.Flush())
	}
	return e.Err
}

func (e *Encoder) EncodeMultiBulk(multi []*Resp, flush bool) error {
	if e.Err != nil {
		return errorsTrace(e.Err)
	}
	if err := e.encodeMultiBulk(multi); err != nil {
		e.Err = err
	} else if flush {
		e.Err = errorsTrace(e.Err)
	}
	return e.Err
}

func (e *Encoder) Flush() error {
	if e.Err != nil {
		return errorsTrace(errorNew("Flush error"))
	}
	if err := e.bw.Flush(); err != nil {
		e.Err = errorsTrace(errorNew("bw.Flush error"))
	}
	return e.Err
}

func (e *Encoder) encodeResp(r *Resp) error {
	if err := e.bw.WriteByte(byte(r.Type)); err != nil {
		return errorsTrace(err)
	}
	switch r.Type {
	case TypeString, TypeError, TypeInt:
		return e.encodeTextBytes(r.Value)
	case TypeBulkBytes:
		return e.encodeBulkBytes(r.Value)
	case TypeArray:
		return e.encodeArray(r.Array)
	default:
		return errorsTrace(e.Err)
	}
}

func (e *Encoder) encodeMultiBulk(multi []*Resp) error {
	if err := e.bw.WriteByte(byte(TypeArray)); err != nil {
		return errorsTrace(err)
	}
	return e.encodeArray(multi)
}

func (e *Encoder) encodeTextBytes(b []byte) error {
	if _, err := e.bw.Write(b); err != nil {
		return errorsTrace(err)
	}
	if _, err := e.bw.WriteString("\r\n"); err != nil {
		return errorsTrace(err)
	}
	return nil
}

func (e *Encoder) encodeTextString(s string) error {
	if _, err := e.bw.WriteString(s); err != nil {
		return errorsTrace(err)
	}
	if _, err := e.bw.WriteString("\r\n"); err != nil {
		return errorsTrace(err)
	}
	return nil
}

func (e *Encoder) encodeInt(v int64) error {
	return e.encodeTextString(strconv.FormatInt(v, 10))
}

func (e *Encoder) encodeBulkBytes(b []byte) error {
	if b == nil {
		return e.encodeInt(-1)
	} else {
		if err := e.encodeInt(int64(len(b))); err != nil {
			return err
		}
		return e.encodeTextBytes(b)
	}
}

func (e *Encoder) encodeArray(array []*Resp) error {
	if array == nil {
		return e.encodeInt(-1)
	} else {
		if err := e.encodeInt(int64(len(array))); err != nil {
			return err
		}
		for _, r := range array {
			if err := e.encodeResp(r); err != nil {
				return err
			}
		}
		return nil
	}
}

// api
func EncodeCmd(cmd string) ([]byte, error) {
	b := []byte(cmd)
	r := bytes.Split(b, []byte(" "))
	if r == nil {
		return nil, errorsTrace(errorNew("empty split"))
	}
	resp := NewArray(nil)
	for _, v := range r {
		if len(v) > 0 {
			resp.Array = append(resp.Array, NewBulkBytes(v))
		}
	}
	return EncodeToBytes(resp)
}

func Encode(w io.Writer, resp *Resp) error {
	return NewEncoder(w).Encode(resp, true)
}

func EncodeToBytes(resp *Resp) ([]byte, error) {
	var b = &bytes.Buffer{}
	if err := Encode(b, resp); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

/*
*
* Decoder 解码器
*
**/
type Decoder struct {
	br *bufio2.Reader

	Err error
}

func NewDecoder(r io.Reader) *Decoder {
	return NewDecoderBuffer(bufio2.NewReaderSize(r, 8192))
}

func NewDecoderSize(r io.Reader, size int) *Decoder {
	return NewDecoderBuffer(bufio2.NewReaderSize(r, size))
}

func NewDecoderBuffer(br *bufio2.Reader) *Decoder {
	return &Decoder{br: br}
}

func (d *Decoder) Decode() (*Resp, error) {
	if d.Err != nil {
		return nil, errorsTrace(errorNew("Decode err"))
	}
	r, err := d.decodeResp()
	if err != nil {
		d.Err = err
	}
	return r, d.Err
}

func (d *Decoder) DecodeMultiBulk() ([]*Resp, error) {
	if d.Err != nil {
		return nil, errorsTrace(errorNew("DecodeMultibulk error"))
	}
	m, err := d.decodeMultiBulk()
	if err != nil {
		d.Err = err
	}
	return m, err
}

func (d *Decoder) decodeResp() (*Resp, error) {
	b, err := d.br.ReadByte()
	if err != nil {
		return nil, errorsTrace(err)
	}
	r := &Resp{}
	r.Type = byte(b)
	switch r.Type {
	default:
		return nil, errorsTrace(err)
	case TypeString, TypeError, TypeInt:
		r.Value, err = d.decodeTextBytes()
	case TypeBulkBytes:
		r.Value, err = d.decodeBulkBytes()
	case TypeArray:
		r.Array, err = d.decodeArray()
	}
	return r, err
}

func (d *Decoder) decodeTextBytes() ([]byte, error) {
	b, err := d.br.ReadBytes('\n')
	if err != nil {
		return nil, errorsTrace(err)
	}
	if n := len(b) - 2; n < 0 || b[n] != '\r' {
		return nil, errorsTrace(err)
	} else {
		return b[:n], nil
	}
}

func (d *Decoder) decodeInt() (int64, error) {
	b, err := d.br.ReadSlice('\n')
	if err != nil {
		return 0, errorsTrace(err)
	}
	if n := len(b) - 2; n < 0 || b[n] != '\r' {
		return 0, errorsTrace(err)
	} else {
		return Btoi64(b[:n])
	}
}

func (d *Decoder) decodeBulkBytes() ([]byte, error) {
	n, err := d.decodeInt()
	if err != nil {
		return nil, err
	}
	switch {
	case n < -1:
		return nil, errorsTrace(err)
	case n > MaxBulkBytesLen:
		return nil, errorsTrace(err)
	case n == -1:
		return nil, nil
	}
	b, err := d.br.ReadFull(int(n) + 2)
	if err != nil {
		return nil, errorsTrace(err)
	}
	if b[n] != '\r' || b[n+1] != '\n' {
		return nil, errorsTrace(err)
	}
	return b[:n], nil
}

func (d *Decoder) decodeArray() ([]*Resp, error) {
	n, err := d.decodeInt()
	if err != nil {
		return nil, err
	}
	switch {
	case n < -1:
		return nil, errorsTrace(err)
	case n > MaxArrayLen:
		return nil, errorsTrace(err)
	case n == -1:
		return nil, nil
	}
	array := make([]*Resp, n)
	for i := range array {
		r, err := d.decodeResp()
		if err != nil {
			return nil, err
		}
		array[i] = r
	}
	return array, nil
}

func (d *Decoder) decodeSingleLineMultiBulk() ([]*Resp, error) {
	b, err := d.decodeTextBytes()
	if err != nil {
		return nil, err
	}
	multi := make([]*Resp, 0, 8)
	for l, r := 0, 0; r <= len(b); r++ {
		if r == len(b) || b[r] == ' ' {
			if l < r {
				multi = append(multi, NewBulkBytes(b[l:r]))
			}
			l = r + 1
		}
	}
	if len(multi) == 0 {
		return nil, errorsTrace(err)
	}
	return multi, nil
}

func (d *Decoder) decodeMultiBulk() ([]*Resp, error) {
	b, err := d.br.PeekByte()
	if err != nil {
		return nil, errorsTrace(err)
	}
	if RespType(b) != TypeArray {
		return d.decodeSingleLineMultiBulk()
	}
	if _, err := d.br.ReadByte(); err != nil {
		return nil, errorsTrace(err)
	}
	n, err := d.decodeInt()
	if err != nil {
		return nil, errorsTrace(err)
	}
	switch {
	case n <= 0:
		return nil, errorsTrace(ErrBadArrayLen)
	case n > MaxArrayLen:
		return nil, errorsTrace(ErrBadArrayLenTooLong)
	}
	multi := make([]*Resp, n)
	for i := range multi {
		r, err := d.decodeResp()
		if err != nil {
			return nil, err
		}
		if r.Type != TypeBulkBytes {
			return nil, errorsTrace(ErrBadMultiBulkContent)
		}
		multi[i] = r
	}
	return multi, nil
}

// api
func Decode(p []byte) (*Resp, error) {
	return NewDecoder(bytes.NewReader(p)).Decode()
}

func DecodeMultiBulk(p []byte) ([]*Resp, error) {
	return NewDecoder(bytes.NewReader(p)).DecodeMultiBulk()
}

/*
*
* Response
*
**/
type Resp struct {
	Type byte

	Value []byte
	Array []*Resp
}

func NewString(value []byte) *Resp {
	r := &Resp{}
	r.Type = TypeString
	r.Value = value
	return r
}

func NewError(value []byte) *Resp {
	r := &Resp{}
	r.Type = TypeError
	r.Value = value
	return r
}

func NewInt(value []byte) *Resp {
	r := &Resp{}
	r.Type = TypeInt
	r.Value = value
	return r
}

// 批量回复类型
func NewBulkBytes(value []byte) *Resp {
	r := &Resp{}
	r.Type = TypeBulkBytes
	r.Value = value
	return r
}

// 多条批量回复类型
func NewArray(array []*Resp) *Resp {
	r := &Resp{}
	r.Type = TypeArray
	r.Array = array
	return r
}

func errorsTrace(err error) error {
	if err != nil {
		log.Println("errors Tracing", err.Error())
	}
	return err
}

func errorNew(msg string) error {
	return errors.New("error occur, msg " + msg)
}