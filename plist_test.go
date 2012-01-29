package plist

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestXMLDefault(t *testing.T) {
	buf, err := Marshal(int64(42))
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

func TestDetectingReaderBinary(t *testing.T) {
	var e Entitlements

	buf := []byte("bplist00......")

	err := Unmarshal(buf, &e)
	if err == nil {
		t.Fatalf("should detect bplist")
	}
}