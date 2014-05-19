package xmlplist

import (
	"bytes"
	"io/ioutil"
	"testing"
)

var mapFiles = []string{
	"testdata/AlfredTimeKeeper.alfredworkflow",
	"testdata/AllowedEmptyArray.plist",
	"testdata/AllowedEmptyData.plist",
	"testdata/AllowedEmptyDict.plist",
	"testdata/AllowedEmptyFields.plist",
	"testdata/AllowedEmptyString.plist",
	"testdata/ConcreteArray.plist",
	"testdata/Date.plist",
	"testdata/Entitlements.plist",
	"testdata/EntitlementsWeirdWhitespace.plist",
	"testdata/MultiTypeArray.plist",
	"testdata/RecursiveEntitlements.plist",
	"testdata/RecursiveStruct.plist.golden",
	"testdata/Struct.plist.golden",
}

var arrayFiles = []string{
	"testdata/Array.plist.golden",
	"testdata/Data.plist.golden",
	"testdata/Date.plist.golden",
	"testdata/DecodeEverything.plist",
	"testdata/DisallowedEmptyDate.plist",
	"testdata/DisallowedEmptyInteger.plist",
	"testdata/DisallowedEmptyReal.plist",
	"testdata/Float.plist.golden",
	"testdata/Int.plist.golden",
	"testdata/String.plist.golden",
}

func TestMapEncoderAndDecoder(t *testing.T) {
	for _, filename := range mapFiles {
		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		var dict map[string]interface{}
		err = Unmarshal(buf, &dict)
		if err != nil {
			t.Fatal(err)
		}
		buf2, err := Marshal(dict)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf, buf2) {
			t.Fatal("Marshal(Unmarshal(x)) != x")
		}
	}
}

func TestArrayEncoderAndDecoder(t *testing.T) {
	for _, filename := range arrayFiles {
		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		var array []interface{}
		err = Unmarshal(buf, &array)
		if err != nil {
			t.Fatal(err)
		}
		buf2, err := Marshal(array)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf, buf2) {
			t.Fatal("Marshal(Unmarshal(x)) != x")
		}
	}
}
