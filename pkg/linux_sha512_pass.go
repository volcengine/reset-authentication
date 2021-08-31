// +build !windows

package pkg

import (
	"bytes"
	"github.com/volcengine/reset-authentication/util"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"time"
)

type LinuxResetPassword struct {
}

func (*LinuxResetPassword) Init() error {
	return nil
}

func (*LinuxResetPassword) GetPassword() (string, error) {
	var url = fmt.Sprintf("http://%s/volcstack/latest/reset_password", gValidDataSource)
	out, err := doDataSourceRequest(url, "GET", "")

	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (*LinuxResetPassword) NeedDecryption() bool {
	return false
}

func (*LinuxResetPassword) Decrypt(password string) (string, error) {
	return password, nil
}

func (*LinuxResetPassword) ResetPassword(password string) error {
	var runCmd = exec.Command("chpasswd", "-e")

	var r, w = io.Pipe()
	defer func() {
		_ = r.Close()
		_ = w.Close()
	}()

	runCmd.Stdin = r

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*30)
	defer cancel()

	var err error
	var errCh = make(chan error, 1)
	go func() {
		var stderr bytes.Buffer
		runCmd.Stderr = &stderr
		if err = runCmd.Start(); err != nil {
			errCh <- fmt.Errorf("StdErr: %s RetErr: %v", string(stderr.Bytes()), err)
		}

		if _, err = w.Write([]byte("root:" + password)); err != nil {
			errCh <- err
		}

		if err = w.Close(); err != nil {
			errCh <- err
		}

		if err = runCmd.Wait(); err != nil {
			errCh <- fmt.Errorf("StdErr: %s RetErr: %v", string(stderr.Bytes()), err)
		}
		errCh <- nil
	}()

	select {
	case err = <-errCh:
		if err != nil {
			return err
		}
		return nil
	case <-ctx.Done():
		if runCmd.Process != nil {
			_ = runCmd.Process.Kill()
		}
		return errors.New("Reset password timeout")
	}
}

func init() {
	if runtime.GOOS == "linux" && util.GetConfig().EnableResetPassword {
		RegisterResetPasswordDriver(&LinuxResetPassword{})
	}
}
