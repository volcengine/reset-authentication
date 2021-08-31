// +build windows

package pkg

import (
	"bytes"
	"github.com/volcengine/reset-authentication/util"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/sys/windows/registry"
)

const (
	uploadKey     = "upload-rsa-pub-done"
	rsaPubkey     = "rsa-key.pub"
	rsaPrvkey     = "rsa-key"
	instanceIdKey = "instance-id"
)

const initDoneValue = "INIT_DONE"

type WindowsResetPassword struct {
	key      registry.Key
	exists   bool
	initDone bool
	pubkey   []byte
}

func (w *WindowsResetPassword) initInstanceId() error {
	id, err := getInstanceId(gValidDataSource)
	if err != nil {
		return err
	}

	if w.exists {
		id, err := getInstanceId(gValidDataSource)
		if err != nil {
			return err
		}

		v, _, err := w.key.GetStringValue(instanceIdKey)
		if err != nil {
			return err
		}

		if v != id {
			_ = w.key.SetStringValue(instanceIdKey, id)
			_ = w.key.SetBinaryValue(rsaPubkey, []byte{})
			_ = w.key.SetBinaryValue(rsaPrvkey, []byte{})
			_ = w.key.SetStringValue(uploadKey, "")
		}
	}

	return w.key.SetStringValue(instanceIdKey, id)
}

func (w *WindowsResetPassword) genRsaKey() ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, nil, err
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	prvkey := pem.EncodeToMemory(block)

	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, nil, err
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	pubkey := pem.EncodeToMemory(block)
	return prvkey, pubkey, nil
}

func (w *WindowsResetPassword) recordRsaKey() error {
	prvKey, pubKey, err := w.genRsaKey()
	if err != nil {
		return err
	}

	if err = w.key.SetBinaryValue(rsaPubkey, pubKey); err != nil {
		return err
	}
	if err = w.key.SetBinaryValue(rsaPrvkey, prvKey); err != nil {
		return err
	}
	w.pubkey = pubKey

	return nil
}

func (w *WindowsResetPassword) uploadPubkey() error {
	var url = fmt.Sprintf("http://%s/volcstack/latest/windows_reset_pass_key", gValidDataSource)
	_, err := doDataSourceRequest(url, "POST", string(w.pubkey))
	if err != nil {
		return err
	}

	if err = w.key.SetStringValue(uploadKey, initDoneValue); err != nil {
		return err
	}
	return nil
}

func (w *WindowsResetPassword) Init() error {
	var retErr = errors.New("Init faild")
	key, exists, err := registry.CreateKey(registry.LOCAL_MACHINE, `SOFTWARE\Reset Authentication`, registry.ALL_ACCESS)
	if err != nil {
		util.Error("Create windows registry failed. Error:", err)
		return retErr
	}

	w.exists = exists
	w.key = key

	// check instanceID has been
	err = w.initInstanceId()
	if err != nil {
		util.Error("Init instance id to registry failed. Error:", err)
		return retErr
	}

	uploadValue, _, _ := w.key.GetStringValue(uploadKey)
	if uploadValue != initDoneValue {
		if err = w.recordRsaKey(); err != nil {
			util.Error("Gen rsa key pair failed. Error:", err)
			return retErr
		}

		if err = w.uploadPubkey(); err != nil {
			util.Error("Upload pubkey to registry failed. Error:", err)
			return retErr
		}
		return nil
	}
	w.initDone = true
	return nil
}

func (w *WindowsResetPassword) GetPassword() (string, error) {
	if !w.initDone {
		util.Info("First init windows. Can not support reset password")
		return "", nil
	}

	var url = fmt.Sprintf("http://%s/volcstack/latest/reset_password", gValidDataSource)
	out, err := doDataSourceRequest(url, "GET", "")

	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (w *WindowsResetPassword) NeedDecryption() bool {
	return true
}

func (w *WindowsResetPassword) Decrypt(password string) (string, error) {
	out, _, err := w.key.GetBinaryValue(rsaPrvkey)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(out)
	if block == nil {
		return "", errors.New("private key error!")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	sDec, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return "", err
	}

	data, err := rsa.DecryptPKCS1v15(rand.Reader, priv, sDec)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (*WindowsResetPassword) ResetPassword(password string) error {
	var runCmd = exec.Command("net", "user", "Administrator", password)
	var errCh = make(chan error, 1)
	var err error

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*30)
	defer cancel()

	go func() {
		var stdout, stderr bytes.Buffer
		runCmd.Stdout = &stdout
		runCmd.Stderr = &stderr
		err := runCmd.Run()
		runCmd.Stderr = &stderr
		if err != nil {
			errCh <- fmt.Errorf("StdErr: %s , StdOut: %s,  RetErr: %v",
				string(stdout.Bytes()), string(stderr.Bytes()), err)
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
	if runtime.GOOS == "windows" {
		RegisterResetPasswordDriver(&WindowsResetPassword{})
	}
}
