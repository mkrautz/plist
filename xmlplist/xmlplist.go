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

// nextElement returns the next StartElement or EndElement
// token found in the stream.
func (d *Decoder) nextElement() (xml.Token, error) {
	for {
		t, err := d.xd.Token()
		if err != nil {
			return nil, err
		}

		switch t.(type) {
		case xml.StartElement, xml.EndElement:
			return t, nil
		}
	}
	return nil, nil
}

// readEndElement reads the end element with the specified name.
// If a non-EndElement is encountered, or a wrong name is encountered
// an error is returned.
func (d *Decoder) readEndElement(name string) error {
	t, err := d.nextElement()
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

// expectWhitespace reads the next element (expected to be a
// charadata token), and checks whether it only contains whitespace.
// If not, it returns an error.
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
			case '\n', '\t', ' ':
				// ok
			default:
				return fmt.Errorf("plist: unexpected character in whitespace: %q", r)
		}
	}
	return nil
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

// expectCharData reads the next token of the stream,
// expecting it to be a CharData token. If it isn't
// an error is returned.
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

// parsePlist parses the first <plist> StartElement and
// begins reading the root element of the plist into v.
func (d *Decoder) parsePlist(v interface{}) error {
	t, err := d.nextElement()
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

	// Read the root element of the plist
	t, err = d.nextElement()
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


// readType reads a root element into v. The type of the root element
// is deducted from the start element passed as se.
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

// readDict parses an XML plist dictionary. The StartElement given
// must be a dict StartElement. readDict fills out the the value v
// with the contents of the dict. readDict ends reading the dictionary
// once it encounters an end element with the same name as was given in
// se.
func (d *Decoder) readDict(v interface{}, se xml.StartElement) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errors.New("plist: v must be ptr")
	}

	dictMap := map[string]interface{}{}
	for {
		// read <key>
		t, err := d.nextElement()
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
		t, err = d.nextElement()
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

		// read type
		t, err = d.nextElement()
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

// readArray reads an XML plist array into v. The se parameter must be
// a StartElement with name array. readArray stops reading when an
// EndElement with name array is encountered.
func (d *Decoder) readArray(v interface{}, se xml.StartElement) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errors.New("plist: v must be ptr")
	}

	var slice []interface{}
	for {
		t, err := d.nextElement()
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

// readBool reads an XML plist boolean into the value v.
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

// readDate reads an XML plist date into the value v.
func (d *Decoder) readDate(v interface{}, se xml.StartElement) error {
	return errors.New("plist: date support not yet implemented")
}

// readData reads an XML plist data blob into the value v.
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

// readString reads an XML plist string into the value v.
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

// readReal reads an XML plist real number into the value v.
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

// readInteger reads an XML plist integer into the value v.
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