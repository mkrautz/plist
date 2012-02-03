package plist

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestXMLDefault(t *testing.T) {
	buf, err := Marshal([]int64{int64(42)})
	if err != nil {
		t.Fatalf("unable to marshal: %v", err)
	}

	expectedBuf, err := ioutil.ReadFile("xmlplist/testdata/Int.plist.golden")
	if err != nil {
		t.Fatalf("unable to read golden int")
	}

	if !bytes.Equal(buf, expectedBuf) {
		t.Fatalf("golden mismatch")
	}
}

type Entitlements struct {
	GetTaskAllow bool `plist:"get-task-allow"`
}

func TestDetectingReaderXML(t *testing.T) {
	var e Entitlements

	buf, err := ioutil.ReadFile("xmlplist/testdata/Entitlements.plist")
	if err != nil {
		t.Fatalf("unable to read entitlements: %v", err)
	}

	err = Unmarshal(buf, &e)
	if err != nil {
		t.Fatalf("unable to unmarshal: %v", err)
	}

	if e.GetTaskAllow != true {
		t.Fatalf("unmarshla failed")
	}
}

func TestDetectingReaderASCII(t *testing.T) {
	var val []interface{}

	buf, err := ioutil.ReadFile("asciiplist/testdata/Array.plist")
	if err != nil {
		t.Fatalf("unable to read array plist: %v", err)
	}

	err = Unmarshal(buf, &val)
	if err != nil {
		t.Fatalf("unable to unmarshal: %v", err)
	}

	expected := []string{"hey", "what", "up"}
	if len(expected) != len(val) {
		t.Fatalf("unexpected read from ascii plist decoder")
	}

	for i, str := range expected {
		valStr, ok := val[i].(string)
		if !ok {
			t.Fatalf("non-string in returned array")
		}
		if valStr != str {
			t.Fatalf("ascii plist mismatch")
		}
	}
}

func TestDetectingReaderBinary(t *testing.T) {
	var e Entitlements

	buf := []byte("bplist00......")

	err := Unmarshal(buf, &e)
	if err == nil {
		t.Fatalf("should detect bplist")
	}
}
