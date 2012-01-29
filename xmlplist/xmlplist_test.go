package xmlplist

import (
	"io/ioutil"
	"testing"
)

func TestUnmarshalEntitlements(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/Entitlements.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var dict map[string]interface{}
	err = Unmarshal(buf, &dict)
	if err != nil {
		t.Fatalf("%v", err)
	}

	getTaskAllow, ok := dict["get-task-allow"].(bool)
	if !ok {
		t.Fatalf("get-task-allow not bool")
	}
	if getTaskAllow != true {
		t.Fatalf("get-task-allow not true")
	}
}

type Entitlements struct {
	GetTaskAllow bool `plist:"get-task-allow"`
}

func TestUnmarshalEntitlementsToStruct(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/Entitlements.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var e Entitlements
	err = Unmarshal(buf, &e)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if e.GetTaskAllow != true {
		t.Fatalf("get-task-allow is false")
	}
}

func TestUnmarshalEntitlementsWeirdWhitespaceToStruct(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/EntitlementsWeirdWhitespace.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var e Entitlements
	err = Unmarshal(buf, &e)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if e.GetTaskAllow != true {
		t.Fatalf("get-task-allow is false")
	}
}

func TestUnmarshalBooleanRootElement(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/BoolRoot.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var b bool
	err = Unmarshal(buf, &b)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if b != false {
		t.Fatalf("t should be false")
	}
}

func TestUnmarshalStringRootElement(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/StringRoot.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var str string
	err = Unmarshal(buf, &str)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if str != "hey what up?" {
		t.Fatalf("str mismatch")
	}
}

func TestUnmarshalRealRootElement(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/RealRoot.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var pi float64
	err = Unmarshal(buf, &pi)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if pi != 3.14159265 {
		t.Fatalf("pi mismatch")
	}
}

func TestUnmarshalDataRootElement(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/DataRoot.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var data []byte
	err = Unmarshal(buf, &data)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if string(data) != "hey what up" {
		t.Fatalf("data mismatch")
	}
}

func TestUnmarshalIntegerRootElement(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/IntegerRoot.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var val int64
	err = Unmarshal(buf, &val)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if val != 424344 {
		t.Fatalf("integer mismatch")
	}
}

func TestUnmarshalArrayRootElement(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/ArrayRoot.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var array []interface{}
	err = Unmarshal(buf, &array)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expected := []string{"hey", "what", "up"}
	if len(expected) != len(array) {
		t.Fatalf("bad len")
	}
	for i, s := range expected {
		if array[i].(string) != s {
			t.Fatalf("bad value at idx=%v", i)
		}
	}
}