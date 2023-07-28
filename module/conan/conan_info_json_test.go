package conan

import (
	"path"
	"runtime"
	"testing"
)

func TestConan1ToTree(t *testing.T) {
	var conanJson _ConanInfoJsonFile
	var conan2Json _Conan2InfoJsonFile
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}

	if e := conanJson.ReadFromFile(path.Join(abPath, "../../test_sample/conan.json")); e != nil {
		t.Error(e)
		return
	}
	tree, e := conanJson.Tree()
	if e != nil {
		t.Error(e)
		return
	}
	t.Log(tree)

	if e := conan2Json.ReadFromFile(path.Join(abPath, "../../test_sample/conan2.json")); e != nil {
		t.Error(e)
		return
	}
	tree2, e := conan2Json.Tree()
	if e != nil {
		t.Error(e)
		return
	}

	t.Log(tree2)
}
