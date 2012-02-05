package asciiplist

import (
	"bytes"
	"errors"
	"io"
	"reflect"
)

func Unmarshal(buf []byte, v interface{}) error {
	dec := NewDecoder(bytes.NewBuffer(buf))
	err := dec.Decode(v)
	if err != nil {
		return err
	}
	return nil
}

type Decoder struct {
	s *scanner
}

func NewDecoder(r io.Reader) *Decoder {
	dec := new(Decoder)
	dec.s = newScanner(r)
	return dec
}

func (d *Decoder) Decode(v interface{}) error {
	tok, err := d.s.Token()
	if err != nil {
		return err
	}
	switch tok.(type) {
	case tokenParenOpen:
		return d.readArray(v)
	case tokenCurlyOpen:
		return d.readDict(v)
	default:
		return errors.New("plist: bad root token found in stream")
	}
	return nil
}

func (d *Decoder) readArray(v interface{}) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errors.New("plist: v must be ptr")
	}

	var slice []interface{}
Loop:
	for {
		tok, err := d.s.Token()
		if err != nil {
			return err
		}

		switch tok.(type) {
		case tokenParenOpen:
			var array []interface{}
			err = d.readArray(&array)
			if err != nil {
				return err
			}
			slice = append(slice, array)
		case tokenCurlyOpen:
			var m map[string]interface{}
			err = d.readDict(&m)
			if err != nil {
				return err
			}
		case tokenString:
			slice = append(slice, string(tok.(tokenString)))
		case tokenData:
			slice = append(slice, []byte(tok.(tokenData)))
		default:
			return errors.New("plist: bad array element token")
		}

		tok, err = d.s.Token()
		if err != nil {
			return err
		}

		switch tok.(type) {
		case tokenComma:
			continue
		case tokenParenClose:
			break Loop
		default:
			return errors.New("plist: expected comma or end paren")
		}
	}

	rv := reflect.ValueOf(v).Elem()
	rv.Set(reflect.ValueOf(slice))

	return nil
}

func (d *Decoder) readDict(v interface{}) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errors.New("plist: v must be ptr")
	}

	m := map[string]interface{}{}
Loop:
	for {
		tok, err := d.s.Token()
		if err != nil {
			return err
		}

		var keyName string
		switch tok.(type) {
		case tokenString:
			keyName = string(tok.(tokenString))
		case tokenCurlyClose:
			break Loop
		default:
			return errors.New("plist: bad dict key")
		}

		tok, err = d.s.Token()
		if err != nil {
			return err
		}

		if _, ok := tok.(tokenEqual); !ok {
			return errors.New("plist: expected equal token")
		}

		tok, err = d.s.Token()
		if err != nil {
			return err
		}

		switch tok.(type) {
		case tokenParenOpen:
			var array []interface{}
			err = d.readArray(&array)
			if err != nil {
				return err
			}
			m[keyName] = array
		case tokenCurlyOpen:
			var dict map[string]interface{}
			err = d.readDict(&dict)
			if err != nil {
				return err
			}
			m[keyName] = dict
		case tokenString:
			m[keyName] = string(tok.(tokenString))
		case tokenData:
			m[keyName] = []byte(tok.(tokenData))
		default:
			return errors.New("plist: bad dict value element token")
		}

		tok, err = d.s.Token()
		if err != nil {
			return err
		}

		if _, ok := tok.(tokenSemi); !ok {
			return errors.New("plist: expected semi token")
		}
	}

	rv := reflect.ValueOf(v).Elem()
	mapToValue(m, rv)

	return nil
}

// mapToStruct converts the map-representation of the dictionary in dict
// into a struct or a map given as val. Recursive structs and maps are
// supported.
func mapToValue(dict map[string]interface{}, val reflect.Value) { 
	if val.Kind() == reflect.Map {
		val.Set(reflect.ValueOf(dict))
	} else if val.Kind() == reflect.Struct {
		typ := val.Type()
		nfields := val.NumField()
		for i := 0; i < nfields; i++ {
			f := typ.Field(i)
			name := f.Tag.Get("plist")
			if name != "" {
				if dictVal, ok := dict[name]; ok {
					fieldVal := val.Field(i)
					if fieldVal.Kind() == reflect.Map || fieldVal.Kind() == reflect.Struct {
						newDict, ok := dictVal.(map[string]interface{})
						if !ok {
							return
						}
						mapToValue(newDict, fieldVal)
					} else {
						if fieldVal.Type() == reflect.ValueOf(dictVal).Type() {
							fieldVal.Set(reflect.ValueOf(dictVal))
						}
					}
				}
			}
		}
	}
}