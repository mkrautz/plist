package xmlplist

import (
	"log"
	"testing"
)

// todo(mkrautz): Implement proper tests, not just prints

func TestIntEncode(t *testing.T) {
	val := int64(42)
	buf, err := Marshal(val)
	if err != nil {
		t.Fatalf("unable to marshal: %v", err)
	}

	log.Printf("%v", string(buf))
}

func TestArrayEncode(t *testing.T) {
	slice := []int64{1, 2, 3}
	buf, err := Marshal(slice)
	if err != nil {
		t.Fatalf("unable to marshal: %v", err)
	}

	log.Printf("%v", string(buf))
}

func TestByteSliceEncode(t *testing.T) {
	bytes := []byte{0xff, 0xff, 0xff}
	buf, err := Marshal(bytes)
	if err != nil {
		t.Fatalf("unable to marshal: %v", err)
	}

	log.Printf("%v", string(buf))
}

func TestFloatEncode(t *testing.T) {
	f := float64(3.14159265)
	buf, err := Marshal(f)
	if err != nil {
		t.Fatalf("unable to marshal: %v", err)
	}

	log.Printf("%v", string(buf))
}

func TestEncodeString(t *testing.T) {
	str := "hey what < is up />"
	buf, err := Marshal(str)
	if err != nil {
		t.Fatalf("unable to marshal: %v", err)
	}

	log.Printf("%v", string(buf))
}

func TestEncodeMap(t *testing.T) {
	dict := map[string]interface{}{
		"one": 1,
		"two": 2,
		"three": 3,
		"bytes": []byte{0xff, 0xe3, 0xaf},
	}
	buf, err := Marshal(dict)
	if err != nil {
		t.Fatalf("unable to marshal: %v", err)
	}

	log.Printf("%v", string(buf))
}

func TestEncodeStruct(t *testing.T) {
	e := Entitlements{
		GetTaskAllow: true,
	}
	buf, err := Marshal(e)
	if err != nil {
		t.Fatalf("unable to marshal: %v", err)
	}

	log.Printf("%v", string(buf))
}

func TestEncodeRecursiveStruct(t *testing.T) {
	re := RecursiveEntitlements{
		GetTaskAllow: true,
		Entitlements: Entitlements{
			GetTaskAllow: true,
		},
	}
	buf, err := Marshal(re)
	if err != nil {
		t.Fatalf("unable to marshal: %v", err)
	}

	log.Printf("%v", string(buf))
}