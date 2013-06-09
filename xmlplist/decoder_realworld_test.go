package xmlplist

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func TestAlfredWorkflow(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/AlfredTimeKeeper.alfredworkflow")
	if err != nil {
		t.Fatalf("%v", err)
	}
	var workflow map[string]interface{}
	err = Unmarshal(buf, &workflow)
	if err != nil {
		t.Fatalf("%v", err)
	}

	v, ok := workflow["bundleid"]
	if !ok {
		t.Fatalf("expected bundleid key, but wasn't found")
	}
	expected := "com.customct.AlfredTimeKeeper"
	if !reflect.DeepEqual(v, expected) {
		t.Errorf("expected bundleid value %v; got %v", expected, v)
	}
}
