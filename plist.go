// Package plist provides encoding and decoding of Apple's plist data exchange format.
package plist

import (
	"bytes"
	"errors"
	"github.com/mkrautz/plist/asciiplist"
	"github.com/mkrautz/plist/xmlplist"
	"io"
	"strings"
)

type plistEncoder interface {
	Encode(v interface{}) error
}

type plistDecoder interface {
	Decode(v interface{}) error
}

// A Kind represents a kind of plist.
// There are three distinct kinds: ASCII, XML and Binary.
type Kind int

const (
	Unknown Kind = iota
	XML     // XML plists are supported for both reading and writing
	ASCII   // ASCII plists are not currently supported
	Binary  // Binary plists are not currently supported
)

// Unmarshal unmarshals a plist into the value v.
// The value v must be a pointer to a type supported
// by the kind of plist presented in the given data.
func Unmarshal(data []byte, v interface{}) error {
	dec := NewDecoder(bytes.NewBuffer(data))
	return dec.Decode(v)
}

// Marshal marshals the value v into a plist.
func Marshal(v interface{}) ([]byte, error) {
	bw := new(bytes.Buffer)
	enc := NewEncoder(bw)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return bw.Bytes(), nil
}

// a detectingReader is a special io.Reader that can
// detect plist kinds.
type detectingReader struct {
	buf []byte
	r   io.Reader
}

// detectKind reads some bytes off the start of the
// detectingReader's reader to determine which plist
// kind the given data belongs to.
func (r *detectingReader) detectKind() (Kind, error) {
	r.buf = make([]byte, 10)
	_, err := io.ReadFull(r.r, r.buf)
	if err != nil {
		return Unknown, err
	}

	str := string(r.buf)
	if strings.Contains(str, "<?xml") {
		return XML, nil
	} else if strings.Contains(str, "bplist") {
		return Binary, nil
	} else if strings.ContainsAny(str, "{(") {
		return ASCII, nil
	}

	return Unknown, errors.New("plist: unknown kind")
}

func (r *detectingReader) Read(p []byte) (int, error) {
	// As long as we haven't consumed buf yet, read from it.
	if len(r.buf) > 0 {
		n := copy(p, r.buf)
		r.buf = r.buf[n:]
		return n, nil
	}
	return r.r.Read(p)
}

// A Decoder represents a plist decoder.
// The decoder automatically detects the kind of the plist
// it is reading.
type Decoder struct {
	dr        *detectingReader
	plistDec  plistDecoder
}

// NewDecoder creates a new Decoder capable of reading any of the
// three plist kinds.
func NewDecoder(r io.Reader) *Decoder {
	d := new(Decoder)
	d.dr = &detectingReader{
		r: r,
	}
	return d
}

// Decode decodes the plist stream from the Decoder
// into the value v.
func (d *Decoder) Decode(v interface{}) error {
	if d.plistDec == nil {
		kind, err := d.dr.detectKind()
		if err != nil {
			return err
		}
		if kind == XML {
			d.plistDec = xmlplist.NewDecoder(d.dr)
		} else if kind == ASCII {
			d.plistDec = asciiplist.NewDecoder(d.dr)
		} else {
			return errors.New("plist: unknown kind")
		}
	}
	return d.plistDec.Decode(v)
}

// An Encoder encodes values to one of the three plist formats.
type Encoder struct {
	plistEnc plistEncoder
}

// NewEncoder creates a new plist encoder that encodes to the XML
// plist kind. To specify the output format, see NewSpecificEncoder.
func NewEncoder(w io.Writer) *Encoder {
	return NewSpecificEncoder(w, XML)
}

// NewEncoder creates a new plist encoder that encodes to the plist
// kind given as Kind.
func NewSpecificEncoder(w io.Writer, kind Kind) *Encoder {
	enc := new(Encoder)
	switch kind {
	case XML:
		enc.plistEnc = xmlplist.NewEncoder(w)
	default:
		return nil
	}
	return enc
}

// Encode encodes the value v into the plist kind
// the Encoder is configured to use.
func (e *Encoder) Encode(v interface{}) error {
	return e.plistEnc.Encode(v)
}
