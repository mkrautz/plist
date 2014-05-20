package xmlplist

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"
)

// Marshal returns the XML plist encoding of v.
func Marshal(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := NewEncoder(buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// An Encoder encodes Go values into
// the XML plist format.
type Encoder struct {
	w           io.Writer
	bw          *bufio.Writer
	indentLevel int
}

// Returns a string that conforms to the current indent level.
// Strings that are output by the encoder should always have the
// output of this function after a newline.
func (e *Encoder) indent() string {
	var b []byte
	for i := 0; i < e.indentLevel; i++ {
		b = append(b, '\t')
	}
	return string(b)
}

// Writes a string (including proper indentation) to the Encoder.
func (e *Encoder) writeString(str string) error {
	_, err := e.bw.WriteString(e.indent() + str)
	if err != nil {
		return err
	}
	return nil
}

// NewEncoder returns a new Encoder capable of encoding XML plists.
func NewEncoder(w io.Writer) *Encoder {
	enc := new(Encoder)
	enc.w = w
	enc.bw = bufio.NewWriter(w)
	return enc
}

// Encode writes the XML plist encoding of v to the encoder's
// writer.
func (e *Encoder) Encode(v interface{}) error {
	err := e.writeString(xml.Header)
	if err != nil {
		return err
	}

	err = e.writeString("<!" + xmlPlistDocType + ">\n")
	if err != nil {
		return err
	}

	err = e.writeString("<plist version=\"" + xmlPlistVersion + "\">\n")
	if err != nil {
		return err
	}

	if _, isData := v.([]byte); isData {
		return errors.New("plist: bad root element: must be dict or map")
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Struct:
		err = e.encodeAny(v)
		if err != nil {
			return err
		}
	default:
		return errors.New("plist: bad root element: must be dict or map")
	}

	err = e.writeString("</plist>\n")
	if err != nil {
		return err
	}

	err = e.bw.Flush()
	if err != nil {
		return err
	}

	return nil
}

// encodeAny encodes any type into its XML plist equivalent.
func (e *Encoder) encodeAny(v interface{}) (err error) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		if v, data := rv.Interface().([]byte); data {
			err = e.encodeData(v)
		} else {
			err = e.encodeArray(rv)
		}
	case reflect.Struct:
		if v, date := rv.Interface().(time.Time); date {
			err = e.encodeDate(v)
		} else {
			err = e.encodeStruct(rv)
		}
	default:
		switch v := v.(type) {
		case int, int8, int16, int32, int64:
			err = e.encodeInt(rv.Int())
		case float32, float64:
			err = e.encodeFloat(rv.Float())
		case bool:
			err = e.encodeBoolean(v)
		case string:
			err = e.encodeString(v)
		case map[string]interface{}:
			err = e.encodeMap(v)
		default:
			return fmt.Errorf("plist: cannot encode %v (type %T)", v, v)
		}
	}
	return err
}

// encodeArray encodes an array type to the XML plist format.
func (e *Encoder) encodeArray(rv reflect.Value) error {
	err := e.writeString("<array>\n")
	if err != nil {
		return err
	}

	e.indentLevel++

	for i := 0; i < rv.Len(); i++ {
		ev := rv.Index(i).Interface()
		err = e.encodeAny(ev)
		if err != nil {
			return err
		}
	}

	e.indentLevel--

	err = e.writeString("</array>\n")
	if err != nil {
		return err
	}

	return nil
}

// encodeInt encodes an integer type to the XML plist format.
func (e *Encoder) encodeInt(val int64) error {
	err := e.writeString("<integer>" + strconv.FormatInt(val, 10) + "</integer>\n")
	if err != nil {
		return err
	}
	return nil
}

// encodeFloat encodes a floating point number to the XML plist format.
func (e *Encoder) encodeFloat(val float64) error {
	err := e.writeString("<real>" + strconv.FormatFloat(val, 'f', -1, 64) + "</real>\n")
	if err != nil {
		return err
	}
	return nil
}

// encodeData encodes a byte slice to the XML plist format.
func (e *Encoder) encodeData(buf []byte) error {
	err := e.writeString("<data>" + base64.StdEncoding.EncodeToString(buf) + "</data>\n")
	if err != nil {
		return err
	}

	return nil
}

// encodeBoolean encodes a bool to the XML plist format.
func (e *Encoder) encodeBoolean(v bool) error {
	str := "<false/>\n"
	if v {
		str = "<true/>\n"
	}
	err := e.writeString(str)
	if err != nil {
		return err
	}
	return nil
}

// encodeString encodes a string to the XML plist format.
func (e *Encoder) encodeString(str string) error {
	_, err := e.bw.WriteString(e.indent() + "<string>")
	if err != nil {
		return err
	}

	xml.Escape(e.bw, []byte(str))

	_, err = e.bw.WriteString("</string>\n")
	if err != nil {
		return err
	}

	return nil
}

// encodeMap encodes a map to an XML plist dict.
func (e *Encoder) encodeMap(dict map[string]interface{}) error {
	err := e.writeString("<dict>\n")
	if err != nil {
		return err
	}

	e.indentLevel++

	for k, v := range dict {
		_, err = e.bw.WriteString(e.indent() + "<key>")
		if err != nil {
			return err
		}
		xml.Escape(e.bw, []byte(k))
		_, err = e.bw.WriteString("</key>\n")
		if err != nil {
			return err
		}

		err = e.encodeAny(v)
		if err != nil {
			return err
		}
	}

	e.indentLevel--

	err = e.writeString("</dict>\n")
	if err != nil {
		return err
	}

	return nil
}

// encodeStruct encodes a struct to an XML plist dict.
func (e *Encoder) encodeStruct(rv reflect.Value) error {
	err := e.writeString("<dict>\n")
	if err != nil {
		return err
	}

	e.indentLevel++

	for i := 0; i < rv.NumField(); i++ {
		rt := rv.Type()
		f := rt.Field(i)
		name := f.Tag.Get("plist")
		if name == "-" {
			continue
		}
		if name == "" {
			name = f.Name
		}

		_, err = e.bw.WriteString(e.indent() + "<key>")
		if err != nil {
			return err
		}
		xml.Escape(e.bw, []byte(name))
		_, err = e.bw.WriteString("</key>\n")
		if err != nil {
			return err
		}

		err = e.encodeAny(rv.Field(i).Interface())
		if err != nil {
			return err
		}
	}

	e.indentLevel--

	err = e.writeString("</dict>\n")
	if err != nil {
		return err
	}

	return nil
}

// encodeDate encodes a time.Timem to XML plist format.
func (e *Encoder) encodeDate(t time.Time) error {
	str := t.UTC().Format(time.RFC3339)

	err := e.writeString("<date>" + str + "</date>\n")
	if err != nil {
		return err
	}

	return nil
}
