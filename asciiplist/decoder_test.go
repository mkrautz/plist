package asciiplist

import (
	"io/ioutil"
	"testing"
)

func TestReadArray(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/Array.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}

	var slice []interface{}
	err = Unmarshal(buf, &slice)
	if err != nil {
		t.Fatalf("%v", err)
	}

	s, ok := slice[0].(string)
	if !ok {
		t.Fatalf("bad value 0")
	}
	if s != "hey" {
		t.Fatalf("bad string content 0")
	}

	s, ok = slice[1].(string)
	if !ok {
		t.Fatalf("bad value 1")
	}

	if s != "what" {
		t.Fatalf("bad string content 1")
	}

	s, ok = slice[2].(string)
	if !ok {
		t.Fatalf("bad value 2")
	}

	if s != "up" {
		t.Fatalf("bad string content 2")
	}
}

type Dict struct {
	Hey  string `plist:"hey"`
	You  string `plist:"you"`
	What string `plist:"what"`
	Up   string `plist:"up"`
}

func TestReadDict(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/Dict.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}

	var dict Dict
	err = Unmarshal(buf, &dict)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if dict.Hey != "1" {
		t.Fatalf("hey != 1")
	}
	if dict.You != "2" {
		t.Fatalf("hey != 2")
	}
	if dict.What != "3" {
		t.Fatalf("hey != 3")
	}
	if dict.Up != "4" {
		t.Fatalf("hey != 4")
	}
}

type SuperDict struct {
	Name string `plist:"name"`
	Dict Dict   `plist:"dict"`
}

func TestReadRecursiveDict(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/RecursiveDict.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}

	var sd SuperDict
	err = Unmarshal(buf, &sd)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if sd.Name != "SuperDict" {
		t.Fatalf("name != SuperDict")
	}
	if sd.Dict.Hey != "1" {
		t.Fatalf("hey != 1")
	}
	if sd.Dict.You != "2" {
		t.Fatalf("hey != 2")
	}
	if sd.Dict.What != "3" {
		t.Fatalf("hey != 3")
	}
	if sd.Dict.Up != "4" {
		t.Fatalf("hey != 4")
	}
}
