package xmlplist

import (
	"io/ioutil"
	"reflect"
	"testing"
)

type ConcreteArray struct {
	Array []int `plist:"Array"`
}

type InterfaceArray struct {
	Array []interface{} `plist:"Array"`
}

func TestConcreteArray(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/ConcreteArray.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var ca ConcreteArray
	err = Unmarshal(buf, &ca)
	if err != nil {
		t.Fatalf("%v", err)
	}

	expected := []int{0, 1, 2}
	if !reflect.DeepEqual(ca.Array, expected) {
		t.Errorf("got %#v, expected %v", ca.Array, expected)
	}
}

func TestInterfaceArray(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/ConcreteArray.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var ia InterfaceArray
	err = Unmarshal(buf, &ia)
	if err != nil {
		t.Fatalf("%v", err)
	}

	expected := []interface{}{int64(0), int64(1), int64(2)}
	if !reflect.DeepEqual(ia.Array, expected) {
		t.Errorf("got %#v, expected %#v", ia.Array, expected)
	}
}

func TestInterfaceMultiTypeArray(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/MultiTypeArray.plist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var ia InterfaceArray
	err = Unmarshal(buf, &ia)
	if err != nil {
		t.Fatalf("%v", err)
	}

	expected := []interface{}{int64(0), string("hello"), []byte("world")}
	if !reflect.DeepEqual(ia.Array, expected) {
		t.Errorf("got %#v, expected %#v", ia.Array, expected)
	}
}