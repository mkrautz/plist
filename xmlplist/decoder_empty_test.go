package xmlplist

import (
	"io/ioutil"
	"testing"
	"time"
)

type EmptyArray struct {
	EmptyArray []string `plist:"EmptyArray"`
}

type EmptyDict struct {
	EmptyDict map[string]interface{} `plist:"EmptyDict"`
}

type EmptyString struct {
	EmptyString string `plist:"EmptyString"`
}

type EmptyData struct {
	EmptyData []byte `plist:"EmptyData"`
}

type EmptyInteger struct {
	EmptyInteger int `plist:"EmptyInteger"`
}

type EmptyReal struct {
	EmptyReal float64 `plist:"EmptyReal"`
}

type EmptyDate struct {
	EmptyDate time.Time `plist:"EmptyDate"`
}

func TestEmptyArray(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/AllowedEmptyArray.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var ea EmptyArray
	ea.EmptyArray = []string{"non-empty"}
	err = Unmarshal(buf, &ea)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(ea.EmptyArray) > 0 {
		t.Fatalf("ea.EmptyArray length non-zero: %v", len(ea.EmptyArray))
	}
}

func TestEmptyDict(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/AllowedEmptyDict.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var ed EmptyDict
	ed.EmptyDict = map[string]interface{}{
		"hello": "world",
	}
	err = Unmarshal(buf, &ed)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(ed.EmptyDict) > 0 {
		t.Fatalf("ed.EmptyDict length non-zero: %v", len(ed.EmptyDict))
	}
}

func TestEmptyString(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/AllowedEmptyString.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var es EmptyString
	es.EmptyString = "non-empy"
	err = Unmarshal(buf, &es)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(es.EmptyString) > 0 {
		t.Fatalf("es.EmptyString length non-zero: %v", len(es.EmptyString))
	}
}

func TestEmptyData(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/AllowedEmptyData.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var ed EmptyData
	ed.EmptyData = []byte("non-empty")
	err = Unmarshal(buf, &ed)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(ed.EmptyData) > 0 {
		t.Fatalf("ed.EmptyData length non-zero: %v", len(ed.EmptyData))
	}
}

func TestEmptyInteger(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/DisallowedEmptyInteger.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var ei EmptyInteger
	err = Unmarshal(buf, &ei)
	if err == nil || err.Error() != "plist: expected chardata" {
		t.Fatalf("expected chardata error")
	}
}

func TestEmptyReal(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/DisallowedEmptyReal.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var er EmptyReal
	err = Unmarshal(buf, &er)
	if err == nil || err.Error() != "plist: expected chardata" {
		t.Fatalf("expected chardata error")
	}
}

func TestEmptyDate(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/DisallowedEmptyDate.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var ed EmptyDate
	err = Unmarshal(buf, &ed)
	if err == nil || err.Error() != "plist: expected chardata" {
		t.Fatalf("expected chardata error")
	}
}