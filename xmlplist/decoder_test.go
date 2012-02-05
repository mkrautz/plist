package xmlplist

import (
	"io/ioutil"
	"testing"
	"time"
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

type RecursiveEntitlements struct {
	GetTaskAllow bool         `plist:"get-task-allow"`
	Entitlements Entitlements `plist:"Entitlements"`
}

func TestUnmarshalRecursiveEntitlements(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/RecursiveEntitlements.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var r RecursiveEntitlements
	err = Unmarshal(buf, &r)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if r.GetTaskAllow != true {
		t.Fatalf("get-task-allow is false")
	}
	if r.Entitlements.GetTaskAllow != true {
		t.Fatalf("not recursive")
	}
}

func TestUnmarshalRecursiveEntitlementsMap(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/RecursiveEntitlements.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var m map[string]interface{}
	err = Unmarshal(buf, &m)
	if err != nil {
		t.Fatalf("%v", err)
	}

	b, ok := m["get-task-allow"].(bool)
	if !ok {
		t.Fatalf("get-task-allow is not bool")
	}
	if b != true {
		t.Fatalf("get-task-allow not true")
	}

	sm, ok := m["Entitlements"].(map[string]interface{})
	if !ok {
		t.Fatalf("unable to get submap")
	}
	b, ok = sm["get-task-allow"].(bool)
	if !ok {
		t.Fatalf("unable to get get-task-allow from submap as bool")
	}
	if b != true {
		t.Fatalf("get-task-allow from submap is not true")
	}
}

func TestUnmarshalDate(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/Date.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var m map[string]interface{}
	err = Unmarshal(buf, &m)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expectedT, err := time.Parse(time.RFC3339, "2012-01-29T13:07:25Z")
	if err != nil {
		t.Fatalf("%v", err)
	}
	readT, ok := m["now"].(time.Time)
	if !ok {
		t.Fatalf("no date in map")
	}
	if !readT.UTC().Equal(expectedT.UTC()) {
		t.Fatalf("date mismatch")
	}
}

func TestUnmarshalEverythingArray(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/DecodeEverything.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var a []interface{}
	err = Unmarshal(buf, &a)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if len(a) != 6 {
		t.Fatalf("len mismatch")
	}

	integer, ok := a[0].(int64)
	if !ok {
		t.Fatalf("integer type mismatch")
	}
	if integer != 42 {
		t.Fatalf("integer value mismatch")
	}

	real, ok := a[1].(float64)
	if !ok {
		t.Fatalf("real type mismatch")
	}
	if real != float64(50.0) {
		t.Fatalf("real value mismatch")
	}

	expectedT, err := time.Parse(time.RFC3339, "2012-01-29T13:07:25Z")
	if err != nil {
		t.Fatalf("%v", err)
	}
	date, ok := a[2].(time.Time)
	if !ok {
		t.Fatalf("date type mismatch")
	}
	if !date.Equal(expectedT) {
		t.Fatalf("date value mismatch")
	}

	data, ok := a[3].([]byte)
	if !ok {
		t.Fatalf("bad data type")
	}
	if len(data) != 3 {
		t.Fatalf("bad data len")
	}
	if data[0] != 0xff && data[1] != 0xff && data[2] != 0xff {
		t.Fatalf("bad data value")
	}

	str, ok := a[4].(string)
	if !ok {
		t.Fatalf("bad string type")
	}
	if str != "hello" {
		t.Fatalf("bad string value")
	}

	dict, ok := a[5].(map[string]interface{})
	if !ok {
		t.Fatalf("bad dict type")
	}
	str, ok = dict["hey"].(string)
	if !ok {
		t.Fatalf("bad dict key type")
	}
	if str != "ok" {
		t.Fatalf("bad dict key value")
	}
}

type DecodeEverythingWrong struct {
	Integer float64 `plist:"integer"`
	Real int64 `plist:"real"`
	Date []byte `plist:"date"`
	Data map[interface{}]string `plist:"data"`
	Hello string `plist:"hello"`

}

func TestDecodeIntoWrongType(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/DecodeEverythingWrong.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var dew DecodeEverythingWrong
	err = Unmarshal(buf, &dew)
	if err != nil {
		t.Fatalf("%v", err)
	}
}