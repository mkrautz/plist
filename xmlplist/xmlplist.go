// Package xmlplist decodes XML plist files
package xmlplist

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// Unmarshal parses the XML-plist data and stores the result
// in the value pointed to by v.
func Unmarshal(data []byte, v interface{}) error {
	dec := NewDecoder(bytes.NewBuffer(data))
	return dec.Decode(v)
}

// A decoder represents a plist reader that reads
// XML-style plists.
type Decoder struct {
	xd  *xml.Decoder
}

// NewDecoder creates a new XML plist reader.
func NewDecoder(r io.Reader) *Decoder {
	d := new(Decoder)
	d.xd = xml.NewDecoder(r)
	return d
}

// Decode decodes a single XML plist from the decoder.
func (d *Decoder) Decode(v interface{}) error {
	t, err := d.xd.Token()
	if err != nil {
		return err
	}

	// <?xml version="1.0" encoding="UTF-8"?>
	pi, ok := t.(xml.ProcInst)
	if !ok {
		return errors.New("plist: expected ProcInst as first element")
	}
	if pi.Target != "xml" {
		return errors.New("plist: expected xml ProcInst")
	}
	if string(pi.Inst) != `version="1.0" encoding="UTF-8"` {
		return errors.New("plist: unexpected xml ProcInst")
	}

	// \n
	err = d.expectWhitespace()
	if err != nil {
		return err
	}

	// <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
	t, err = d.xd.Token()
	if err != nil {
		return err
	}
	directive, ok := t.(xml.Directive)
	if !ok {
		return errors.New("plist: expected directive")
	}
	if string(directive) != `DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd"` {
		return errors.New("plist: expected plist DTD")
	}

	// \n
	err = d.expectWhitespace()
	if err != nil {
		return err
	}

	return d.parsePlist(v)
}

func (d *Decoder) expectWhitespace() error {
	t, err := d.xd.Token()
	if err != nil {
		return err
	}
	cd, ok := t.(xml.CharData)
	if !ok {
		return errors.New("plist: expected newline")
	}
	for _, r := range cd {
		switch r {
			case '\n':
			case '\t':
			case ' ':
				// ok
			default:
				return fmt.Errorf("plist: unexpected character in whitespace: %q", r)
		}
	}
	return nil
}

func (d *Decoder) expectCharData() ([]byte, error) {
	t, err := d.xd.Token()
	if err != nil {
		return nil, err
	}
	cd, ok := t.(xml.CharData)
	if !ok {
		return nil, errors.New("plist: expected chardata")
	}
	return []byte(cd), nil
}

func (d *Decoder) parsePlist(v interface{}) error {
	t, err := d.xd.Token()
	if err != nil {
		return err
	}

	// <plist version="1.0">
	se, ok := t.(xml.StartElement)
	if !ok {
		return errors.New("plist: expected StartElement")
	}
	if se.Name.Local != "plist" {
		return errors.New("plist: expected <plist> StartElement")
	}
	if len(se.Attr) != 1 {
		return errors.New("plist: unexpected amount of attrs to plist StartElement")
	}
	if se.Attr[0].Name.Local != "version" && se.Attr[0].Value != "1.0" {
		return errors.New("plist: unexpected plist version")
	}

	// \n
	err = d.expectWhitespace()
	if err != nil {
		return err
	}

	// Read the root element of the plist
	t, err = d.xd.Token()
	if err != nil {
		return err
	}

	// If this is indeed a StartElement, what we've got here
	// is the root element. If it isn't, check whether it's an
	// EndElement. It could potentially be the </plist> tag,
	// resulting in an empty plist.
	se, ok = t.(xml.StartElement)
	if !ok {
		if ee, ok := t.(xml.EndElement); ok {
			if ee.Name.Local == "plist" {
				return nil
			}
		}
		return errors.New("plist: expected StartElement (or EndElement)")
	}

	err = d.readType(v, se)
	if err != nil {
		return err
	}

	return nil
}

func (d *Decoder) readType(v interface{}, se xml.StartElement) error {
	switch se.Name.Local {
	case "dict":
		return d.readDict(v, se)
	case "array":
		return d.readArray(v, se)
	case "true", "false":
		return d.readBool(v, se)
	case "date":
		return d.readDate(v, se)
	case "data":
		return d.readData(v, se)
	case "string":
		return d.readString(v, se)
	case "real":
		return d.readReal(v, se)
	case "integer":
		return d.readInteger(v, se)
	}

	return nil
}

func (d *Decoder) readDict(v interface{}, se xml.StartElement) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errors.New("plist: v must be ptr")
	}

	dictMap := map[string]interface{}{}
	for {
		err := d.expectWhitespace()
		if err != nil {
			return err
		}
	
		// read <key>
		t, err := d.xd.Token()
		if err != nil {
			return err
		}
		se, ok := t.(xml.StartElement)
		if !ok {
			if ee, ok := t.(xml.EndElement); ok {
				if ee.Name.Local == "dict" {
					break
				}
				return errors.New("plist: unexpected EndElement")
			}
			return errors.New("plist: expected StartElement")
		}
		if se.Name.Local != "key" {
			return errors.New("bad key name")
		}

		// read key name
		keyNameBuf, err := d.expectCharData()
		if err != nil {
			return err
		}
		keyName := string(keyNameBuf)

		// read </key>
		t, err = d.xd.Token()
		if err != nil {
			return err
		}
		ee, ok := t.(xml.EndElement)
		if !ok {
			return errors.New("plist: unexpected tag")
		}
		if ee.Name.Local != "key" {
			return errors.New("plist: expected end element for key")
		}

		err = d.expectWhitespace()
		if err != nil {
			return err
		}

		// read type
		t, err = d.xd.Token()
		if err != nil {
			return err
		}
		se, ok = t.(xml.StartElement)
		if !ok {
			return errors.New("plist: expected start of type")
		}

		switch se.Name.Local {
		case "dict":
			var m map[string]interface{}
			err = d.readDict(&m, se)
			if err != nil {
				return err
			}
			dictMap[keyName] = m
		case "array":
			var a []interface{}
			err = d.readArray(&a, se)
			if err != nil {
				return err
			}
			dictMap[keyName] = a
		case "true", "false":
			var b bool
			err = d.readBool(&b, se)
			if err != nil {
				return err
			}
			dictMap[keyName] = b
		case "date":
			return errors.New("plist: no date support")
		case "data":
			var buf []byte
			err = d.readData(&buf, se)
			if err != nil {
				return err
			}
			dictMap[keyName] = buf
		case "string":
			var s string
			err = d.readString(&s, se)
			if err != nil {
				return err
			}
			dictMap[keyName] = s
		case "real":
			var f float64
			err = d.readReal(&f, se)
			if err != nil {
				return err
			}
			dictMap[keyName] = f
		case "integer":
			var i int64
			err =  d.readInteger(&i, se)
			if err != nil {
				return err
			}
			dictMap[keyName] = i
		}
	}

	rv := reflect.ValueOf(v).Elem()
	if rv.Kind() == reflect.Map {
		rv.Set(reflect.ValueOf(dictMap))
	} else if rv.Kind() == reflect.Struct {
		rt := rv.Type()
		nfields := rv.NumField()
		for i := 0; i < nfields; i++ {
			f := rt.Field(i)
			name := f.Tag.Get("plist")

			if name != "" {
				if val, ok := dictMap[name]; ok {
					rv.Field(i).Set(reflect.ValueOf(val))
				}
			}
		}
	}

	return nil
}

func (d *Decoder) readArray(v interface{}, se xml.StartElement) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errors.New("plist: v must be ptr")
	}

	var slice []interface{}
	for {
		err := d.expectWhitespace()
		if err != nil {
			return err
		}

		t, err := d.xd.Token()
		if err != nil {
			return err
		}

		se, ok := t.(xml.StartElement)
		if !ok {
			if ee, ok := t.(xml.EndElement); ok {
				if ee.Name.Local == "array" {
					break
				}
				return errors.New("plist: unexepcted EndElement")
			}
			return errors.New("plist: expected StartElement")
		}

		switch se.Name.Local {
		case "dict":
			var m map[string]interface{}
			err = d.readDict(&m, se)
			if err != nil {
				return err
			}
			slice = append(slice, m)
		case "array":
			var a []interface{}
			err = d.readArray(&a, se)
			if err != nil {
				return err
			}
			slice = append(slice, a)
		case "true", "false":
			var b bool
			err = d.readBool(&b, se)
			if err != nil {
				return err
			}
			slice = append(slice, b)
		case "date":
			return errors.New("plist: no date support")
		case "data":
			var buf []byte
			err = d.readData(&buf, se)
			if err != nil {
				return err
			}
			slice = append(slice, buf)
		case "string":
			var s string
			err = d.readString(&s, se)
			if err != nil {
				return err
			}
			slice = append(slice, s)
		case "real":
			var f float64
			err = d.readReal(&f, se)
			if err != nil {
				return err
			}
		case "integer":
			var i int64
			err =  d.readInteger(&i, se)
			if err != nil {
				return err
			}
			slice = append(slice, i)
		}
	}

	rv := reflect.ValueOf(v).Elem()
	rv.Set(reflect.ValueOf(slice))

	return nil
}

func (d *Decoder) readBool(v interface{}, se xml.StartElement) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errors.New("plist: v must be ptr")
	}

	rv := reflect.ValueOf(v).Elem()
	switch se.Name.Local {
	case "true":
		rv.SetBool(true)
		err := d.readEndElement("true")
		if err != nil {
			return err
		}
	case "false":
		rv.SetBool(false)
		err := d.readEndElement("false")
		if err != nil {
			return err
		}
	}

	return nil	
}

func (d *Decoder) readDate(v interface{}, se xml.StartElement) error {
	return errors.New("plist: date support not yet implemented")
}

func (d *Decoder) readData(v interface{}, se xml.StartElement) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errors.New("plist: v must be ptr")
	}

	src, err := d.expectCharData()
	if err != nil {
		return err
	}

	dst, err := base64.StdEncoding.DecodeString(string(src))
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v).Elem()
	rv.Set(reflect.ValueOf(dst))

	err = d.readEndElement("data")
	if err != nil {
		return err
	}

	return nil
}

func (d *Decoder) readString(v interface{}, se xml.StartElement) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errors.New("plist: v must be ptr")
	}

	buf, err := d.expectCharData()
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v).Elem()
	rv.SetString(string(buf))

	err = d.readEndElement("string")
	if err != nil {
		return err
	}

	return nil
}

func (d *Decoder) readReal(v interface{}, se xml.StartElement) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errors.New("plist: v must be ptr")
	}

	buf, err := d.expectCharData()
	if err != nil {
		return err
	}

	str := string(buf)
	rv := reflect.ValueOf(v).Elem()
	var bits int
	switch rv.Kind() {
	case reflect.Float32:
		bits = 32
	case reflect.Float64:
		bits = 64
	}

	f, err := strconv.ParseFloat(str, bits)
	if err != nil {
		return err
	}
	rv.SetFloat(f)

	err = d.readEndElement("real")
	if err != nil {
		return err
	}

	return nil
}

func (d *Decoder) readInteger(v interface{}, se xml.StartElement) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errors.New("plist: v must be ptr")
	}

	buf, err := d.expectCharData()
	if err != nil {
		return err
	}

	str := string(buf)
	rv := reflect.ValueOf(v).Elem()
	var bits int
	switch rv.Kind() {
	case reflect.Int64:
	case reflect.Uint64:
		bits = 64
	case reflect.Int32:
	case reflect.Uint32:
		bits = 32
	case reflect.Int16:
	case reflect.Uint16:
		bits = 16
	case reflect.Int8:
	case reflect.Uint8:
		bits = 8
	}

	val, err := strconv.ParseInt(str, 10, bits)
	if err != nil {
		return err
	}

	rv.SetInt(val)

	err = d.readEndElement("integer")
	if err != nil {
		return err
	}

	return nil
}

func (d *Decoder) readEndElement(name string) error {
	t, err := d.xd.Token()
	if err != nil {
		return err
	}
	ee, ok := t.(xml.EndElement)
	if !ok {
		return err
	}
	if ee.Name.Local != name {
		return fmt.Errorf("plist: expected end element %q", name)
	}
	return nil
}