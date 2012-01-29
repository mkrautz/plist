package asciiplist

import (
	"errors"
	"fmt"
	"io"
)

// See Property List Programming Guide, Appendix A: Old-Style ASCII Property Lists
// (http://developer.apple.com/library/mac/documentation/Cocoa/Conceptual/PropertyLists/PropertyLists.pdf)

type token interface{}
type tokenString string
type tokenData []byte
type tokenParenOpen string
type tokenParenClose string
type tokenCurlyOpen string
type tokenCurlyClose string
type tokenComma string
type tokenSemi string
type tokenEqual string

type scanner struct {
	extra []byte
	r     io.Reader
}

func isAsciiAlphaNumeric(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func hexCharVal(c byte) byte {
	if c >= 'A' && c <= 'F' {
		return 10 + c - 'A'
	} else if c >= 'a' && c <= 'f' {
		return 10 + c - 'a'
	}
	return c - '0'
}

func newScanner(r io.Reader) *scanner {
	return &scanner{r: r}
}

func (s *scanner) getch() (byte, error) {
	if len(s.extra) > 0 {
		c := s.extra[len(s.extra)-1]
		s.extra = s.extra[:len(s.extra)-1]
		return c, nil
	}

	c := make([]byte, 1)
	_, err := s.r.Read(c)
	if err != nil {
		return 0, err
	}

	return c[0], nil
}

func (s *scanner) putch(c byte) {
	s.extra = append(s.extra, c)
}

func (s *scanner) Token() (token, error) {
	for {
		c, err := s.getch()
		if err != nil {
			return nil, err
		}

		switch c {
		case '{':
			return tokenCurlyOpen(c), nil
		case '}':
			return tokenCurlyClose(c), nil
		case '(':
			return tokenParenOpen(c), nil
		case ')':
			return tokenParenClose(c), nil
		case ';':
			return tokenSemi(c), nil
		case ',':
			return tokenComma(c), nil
		case '=':
			return tokenEqual(c), nil
		case ' ', '\t', '\n':
			continue
		case '"':
			return s.scanQuotedString(c)
		case '<':
			return s.scanData(c)
		default:
			if isAsciiAlphaNumeric(c) {
				return s.scanString(c)
			}
			return nil, errors.New("scanner: bad character encountered")
		}
	}

	panic("unreachable")
}

func (s *scanner) scanData(c byte) (token, error) {
	buf := []byte{}

	for {
		c, err := s.getch()
		if err != nil {
			return nil, err
		}

		if c == '>' {
			if len(buf)%2 != 0 {
				return nil, errors.New("scanner: bad data")
			}
			data := []byte{}
			for len(buf) > 0 {
				b := buf[:2]
				data = append(data, hexCharVal(b[0])*16+hexCharVal(b[1]))
				buf = buf[2:]
			}
			return tokenData(data), nil
		} else if c == ' ' {
			continue
		} else {
			if (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') || (c >= '0' && c <= '9') {
				buf = append(buf, c)
			} else {
				return nil, fmt.Errorf("scanner: non-hex ascii character found in data: %v", c)
			}
		}
	}

	panic("unreachable")
}

func (s *scanner) scanQuotedString(c byte) (token, error) {
	buf := []byte{}

	for {
		c, err := s.getch()
		if err != nil {
			return nil, err
		}

		if c == '"' {
			return tokenString(buf), nil
		} else if c == '\\' {
			c, err = s.getch()
			if err != nil {
				return nil, err
			}
			if c == '"' {
				buf = append(buf, c)
			} else {
				return nil, errors.New("scanner: bad escape sequence")
			}
		} else {
			buf = append(buf, c)
		}
	}

	panic("unreachable")
}

func (s *scanner) scanString(c byte) (token, error) {
	buf := []byte{c}

	for {
		c, err := s.getch()
		if err != nil {
			return nil, err
		}

		if !isAsciiAlphaNumeric(c) {
			s.putch(c)
			return tokenString(buf), nil
		} else {
			buf = append(buf, c)
		}
	}

	panic("unreachable")
}
