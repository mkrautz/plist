package xmlplist

import (
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
	//"testdata/DisallowedEmptyDate.plist",
	//"testdata/DisallowedEmptyInteger.plist",
	//"testdata/DisallowedEmptyReal.plist",
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
		var dict1 map[string]interface{}
		err = Unmarshal(buf, &dict1)
		if err != nil {
			t.Fatal(err)
		}
		buf2, err := Marshal(dict1)
		if err != nil {
			t.Fatal(err)
		}
		var dict2 map[string]interface{}
		err = Unmarshal(buf2, &dict2)
		if err != nil {
			t.Fatal(err)
		}
		// Make sure the resulting dict has all the same keys
		// as the original. Checking equality of the values
		// would be a much bigger pain.
		for k := range dict1 {
			if _, ok := dict2[k]; !ok {
				t.Fatal("Unmarshal(Marshal(Unmarshal(x))) != Marshal(x)")
			}
		}
	}
}

func TestArrayEncoderAndDecoder(t *testing.T) {
	for _, filename := range arrayFiles {
		buf1, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		var array1 []interface{}
		err = Unmarshal(buf1, &array1)
		if err != nil {
			t.Fatal(err)
		}
		buf2, err := Marshal(array1)
		if err != nil {
			t.Fatal(err)
		}
		var array2 []interface{}
		err = Unmarshal(buf2, &array2)
		if err != nil {
			t.Fatal(err)
		}
		// Same as above, checking the equality of the values
		// would be an undertaking.
		if len(array1) != len(array2) {
			t.Fatal("Unmarshal(Marshal(Unmarshal(x))) != Marshal(x)")
		}
	}
}
