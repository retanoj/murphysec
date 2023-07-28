package conan

import (
	"context"
	"fmt"
	"github.com/murphysecurity/murphysec/utils/must"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"testing"
)

func TestGetConanVersion(t *testing.T) {
	_, e := exec.LookPath("conan")
	if os.Getenv("CI") != "" && errors.Is(e, exec.ErrNotFound) {
		t.Skip("Conan not found in CI environment, test skipped.")
		return
	}
	t.Log(GetConanVersion(context.TODO(), must.A(LocateConan(context.TODO()))))
}

func TestGetConanInfo(t *testing.T) {
	conanInfo, e := getConanInfo(context.TODO())
	if e != nil {
		t.Error(e)
	}

	t.Log(conanInfo)
	t.Log(fmt.Sprintf("MajorVersion: %d", conanInfo.MajorVersion()))
}
