package asciiplist

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

type scannerTest struct {
	Text     string
	Expected []token
}

var scannerTests []scannerTest = []scannerTest{
	{
		"{ hey = 1; }",
		[]token{
			tokenCurlyOpen("{"),
			tokenString("hey"),
			tokenEqual("="),
			tokenString("1"),
			tokenSemi(";"),
			tokenCurlyClose("}"),
		},
	},
	{
		`{ "hey" = 4; }`,
		[]token{
			tokenCurlyOpen("{"),
			tokenString("hey"),
			tokenEqual("="),
			tokenString("4"),
			tokenSemi(";"),
			tokenCurlyClose("}"),
		},
	},
	{
		`{ "hey\"" = 42; }`,
		[]token{
			tokenCurlyOpen("{"),
			tokenString(`hey"`),
			tokenEqual("="),
			tokenString("42"),
			tokenSemi(";"),
			tokenCurlyClose("}"),
		},
	},
	{
		`("San Francisco", "New York", "Seoul", "London", "Seattle", "Shanghai")`,
		[]token{
			tokenParenOpen("("),
			tokenString("San Francisco"),
			tokenComma(","),
			tokenString("New York"),
			tokenComma(","),
			tokenString("Seoul"),
			tokenComma(","),
			tokenString("London"),
			tokenComma(","),
			tokenString("Seattle"),
			tokenComma(","),
			tokenString("Shanghai"),
			tokenParenClose(")"),
		},
	},
	{
		`{ blob = <cafe babe>; }`,
		[]token{
			tokenCurlyOpen("{"),
			tokenString("blob"),
			tokenEqual("="),
			tokenData([]byte{0xca, 0xfe, 0xba, 0xbe}),
			tokenSemi(";"),
			tokenCurlyClose("}"),
		},
	},
}

func TestScanner(t *testing.T) {
	for _, st := range scannerTests {
		bw := bytes.NewBufferString(st.Text)
		actual := []token{}
		s := newScanner(bw)
		for {
			tok, err := s.Token()
			if err == io.EOF {
				break
			} else if err != nil {
				t.Fatalf("%v", err)
			}
			actual = append(actual, tok)
		}
		if len(st.Expected) != len(actual) {
			t.Fatalf("len mismatch")
		}
		for i := 0; i < len(st.Expected); i++ {

			if !reflect.DeepEqual(st.Expected[i], actual[i]) {
				t.Fatalf("unexpected token: %v", actual[i])
			}
		}
	}
}
