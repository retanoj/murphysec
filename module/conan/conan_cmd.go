package conan

import (
	"context"
	"fmt"
	"github.com/murphysecurity/murphysec/errors"
	"github.com/murphysecurity/murphysec/utils"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type CmdInfo struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

func (c CmdInfo) String() string {
	return fmt.Sprintf("[%s]%s", c.Version, c.Path)
}

func (c CmdInfo) MajorVersion() uint8 {
	i, _ := strconv.Atoi(string(c.Version[0]))
	return uint8(i)
}

var _conanCmdInfo any

func getConanInfo(ctx context.Context) (*CmdInfo, error) {
	if info, ok := _conanCmdInfo.(*CmdInfo); ok {
		return info, nil
	}
	if e, ok := _conanCmdInfo.(error); ok {
		return nil, e
	}
	p, e := LocateConan(ctx)
	if e != nil {
		_conanCmdInfo = e
		return nil, e
	}
	v, e := GetConanVersion(ctx, p)
	if e != nil {
		_conanCmdInfo = e
		return nil, e
	}
	r := &CmdInfo{
		Path:    p,
		Version: v,
	}
	_conanCmdInfo = r
	return r, nil
}

func ExecuteConanInfoCmd(ctx context.Context, cmdInfo *CmdInfo, dir string) (string, error) {
	logger := utils.UseLogger(ctx)
	lp := utils.NewLogPipe(logger, "conan")
	defer lp.Close()
	jsonP := getConanInfoJsonPath()
	logger.Sugar().Debugf("temp file: %s", jsonP)

	var c *exec.Cmd

	sb := utils.MkSuffixBuffer(1024)
	logPipe := utils.NewLogPipe(logger, "conan")
	defer logPipe.Close()

	switch cmdInfo.MajorVersion() {
	case 1:
		c = exec.Command(cmdInfo.Path, "info", ".", "-j", jsonP)
		c.Stdout = io.MultiWriter(sb, logPipe)
		c.Stderr = io.MultiWriter(sb, logPipe)
	case 2:
		c = exec.Command(cmdInfo.Path, "graph", "info", ".", "-f", "json")
		f, _ := os.OpenFile(jsonP, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		defer f.Close()
		c.Stdout = f
		c.Stderr = io.MultiWriter(sb, logPipe)
	default:
		conanErr := conanError("not support conan version")
		return "", conanErr
	}

	logger.Sugar().Infof("Command: %s", c.String())
	c.Env = getEnvForConan()
	c.Dir = dir
	if e := c.Run(); e != nil {
		logger.Warn("Conan command exit with error", zap.Error(e))
		conanErr := conanError(sb.Bytes())
		return "", conanErr
	}
	logger.Info("Conan command completed")
	return jsonP, nil
}

func getConanInfoJsonPath() string {
	return filepath.Join(os.TempDir(), fmt.Sprint("conan-info-", rand.Uint64(), ".json"))
}

func LocateConan(ctx context.Context) (string, error) {
	s, e := exec.LookPath("conan")
	if e != nil {
		return "", errors.WithCause(ErrConanNotFound, e)
	}
	return s, nil
}

func GetConanVersion(ctx context.Context, conanPath string) (string, error) {
	c := exec.CommandContext(ctx, conanPath, "-v")
	c.Env = getEnvForConan()
	if data, e := c.Output(); e != nil {
		return "", errors.WithCause(ErrGetConanVersionFail, e)
	} else {
		return strings.TrimSpace(strings.TrimPrefix(string(data), "Conan version")), nil
	}
}

func getEnvForConan() []string {
	osEnv := os.Environ()
	var rs = make([]string, 0, len(osEnv)+3)
	rs = append(rs, osEnv...)
	return append(rs, "CONAN_NON_INTERACTIVE=1", "NO_COLOR=1", "CLICOLOR=0")
}
