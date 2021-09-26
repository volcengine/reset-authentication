// +build !windows

package pkg

import (
	"bufio"
	"bytes"
	"github.com/volcengine/reset-authentication/util"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type LinuxResetPublicKey struct {
}

func (*LinuxResetPublicKey) CleanPassword() {
	var runCmd = exec.Command("passwd", "-d", "root")

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*30)
	defer cancel()

	var err error
	var errCh = make(chan error)

	go func() {
		var stderr bytes.Buffer
		runCmd.Stderr = &stderr
		err := runCmd.Run()

		if err != nil {
			errCh <- fmt.Errorf("StdErr: %s RetErr: %v", string(stderr.Bytes()), err)
		}
	}()

	select {
	case err = <-errCh:
		if err != nil {
			util.Info("Clean password success")
		}
		util.Info("Clean password failed:", err)
	case <-ctx.Done():
		if runCmd.Process != nil {
			_ = runCmd.Process.Kill()
		}
		util.Info("Clean password timeout")
	}
}

func (*LinuxResetPublicKey) GetDelPublicKey() (string, error) {
	var url = fmt.Sprintf("http://%s/volcstack/latest/reset_del_pubkey", gValidDataSource)
	out, err := doDataSourceRequest(url, "GET", "")

	if err != nil {
		return "", err
	}

	util.Info("Get del public key success:", string(out))
	return string(out), nil
}

func (*LinuxResetPublicKey) GetAddPublicKey() (string, error) {
	var url = fmt.Sprintf("http://%s/volcstack/latest/reset_add_pubkey", gValidDataSource)
	out, err := doDataSourceRequest(url, "GET", "")

	if err != nil {
		return "", err
	}

	util.Info("Get add public key success:", string(out))
	return string(out), nil
}

func (l *LinuxResetPublicKey) ResetPublicKey(delKey, addKey string) error {
	const sshAuthKeyFile = "/root/.ssh/authorized_keys"

	if delKey == "" && addKey == "" {
		return NoPublicKey
	}

	if delKey == addKey {
		util.Info("The delKey add addKey is same")
		return NoPublicKey
	}

	// Create /root/.ssh if /root/.ssh/ is not exist.
	dir := filepath.Dir(sshAuthKeyFile)
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(dir, 0700)
			if err != nil {
				return err
			}
			util.Info("Create dir /root/.ssh success.")
		} else {
			return err
		}
	}

	file, err := os.OpenFile(sshAuthKeyFile, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		// ignore this error
		util.Error("Flock", sshAuthKeyFile, "file failed. Error:", err)
	} else {
		defer func() {
			_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
		}()
	}

	newAuthKeys := make([]string, 0)
	findAddKey := false
	br := bufio.NewReader(file)
	for {
		line, _, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		if string(line) != delKey {
			newAuthKeys = append(newAuthKeys, string(line))
		}

		if addKey != "" {
			if string(line) == addKey {
				findAddKey = true
			}
		}
	}

	if !findAddKey {
		newAuthKeys = append(newAuthKeys, addKey)
	}

	// Truncate authorized_keys files
	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}
	err = file.Truncate(0)
	if err != nil {
		return err
	}

	value := strings.Join(newAuthKeys, "\n")
	if value != "" {
		value += "\n"
	}

	_, err = file.WriteString(value)
	if err != nil {
		return err
	}

	_ = file.Sync()

	util.Info("Reset public key success.")

	if addKey != "" {
		l.CleanPassword()
	}

	return nil
}

func init() {
	if runtime.GOOS == "linux" && util.GetConfig().EnableResetPublicKey {
		RegisterResetSshPublicKeyDriver(&LinuxResetPublicKey{})
	}
}
