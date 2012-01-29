package xmlplist

import (
	"bytes"
	"io/ioutil"
	"testing"
	"time"
)

type EncoderTest struct {
	GoldenFile  string
	Value       interface{}
}

func onceUponATime() time.Time {
	t, err := time.Parse(time.RFC3339, "2012-01-29T13:07:25Z")
	if err != nil {
		return t
	}
	return t.UTC()
}

var encTests []EncoderTest = []EncoderTest{
	{
		"testdata/Int.plist.golden",
		int64(42),
	},
	{
		"testdata/Array.plist.golden",
		[]int64{1, 2, 3},
	},
	{
		"testdata/Data.plist.golden",
		[]byte{0xff, 0xff, 0xff},
	},
	{
		"testdata/Float.plist.golden",
		float64(3.14159265),
	},
	{
		"testdata/String.plist.golden",
		"hey what < is up />",
	},
	{
		"testdata/Struct.plist.golden",
		Entitlements{
			GetTaskAllow: true,
		},
	},
	{
		"testdata/RecursiveStruct.plist.golden",
		RecursiveEntitlements{
			GetTaskAllow: true,
			Entitlements: Entitlements{
				GetTaskAllow: true,
			},
		},
	},
	{
		"testdata/Date.plist.golden",
		onceUponATime(),
	},
}

func TestEncoder(t *testing.T) {
	for _, test := range encTests {
		buf, err := ioutil.ReadFile(test.GoldenFile)
		if err != nil {
			t.Fatalf("%v", err)
		}

		bw := new(bytes.Buffer)
		e := NewEncoder(bw)
		err = e.Encode(test.Value)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if !bytes.Equal(bw.Bytes(), buf) {
			t.Fatalf("test mismatch for file: %v", test.GoldenFile)
		}
	}
}