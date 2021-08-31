package pkg

import (
	"errors"
)

var (
	NoResetPassDriver   = errors.New("ResetPasswordDriver is not implemented")
	NoResetPubkeyDriver = errors.New("ResetSshPublicKeyDriver is not implemented")
	NoPassword          = errors.New("No password needs to be set")
	NoPublicKey         = errors.New("No public key needs to be set")
	NoValidDataSource   = errors.New("No valid datasource found")
)

type ResetPasswordDriver interface {
	Init() error
	GetPassword() (string, error)
	NeedDecryption() bool
	Decrypt(password string) (string, error)
	ResetPassword(password string) error
}

type ResetSshPublicKeyDriver interface {
	GetDelPublicKey() (string, error)
	GetAddPublicKey() (string, error)
	ResetPublicKey(delKey, addKey string) error
}

var gResetPasswordDriver ResetPasswordDriver = nil
var gResetSshPublicKeyDriver ResetSshPublicKeyDriver = nil

func RegisterResetPasswordDriver(driver ResetPasswordDriver) {
	gResetPasswordDriver = driver
}

func RegisterResetSshPublicKeyDriver(driver ResetSshPublicKeyDriver) {
	gResetSshPublicKeyDriver = driver
}

func ResetPassword() error {
	if gResetPasswordDriver == nil {
		return NoResetPassDriver
	}

	err := gResetPasswordDriver.Init()
	if err != nil {
		return err
	}

	password, err := gResetPasswordDriver.GetPassword()
	if err != nil {
		return err
	}

	if password == "" {
		return NoPassword
	}

	if gResetPasswordDriver.NeedDecryption() {
		password, err = gResetPasswordDriver.Decrypt(password)
		if err != nil {
			return err
		}
	}

	return gResetPasswordDriver.ResetPassword(password)
}

func ResetSshPublicKey() error {
	if gResetSshPublicKeyDriver == nil {
		return NoResetPubkeyDriver
	}

	delKey, err := gResetSshPublicKeyDriver.GetDelPublicKey()
	if err != nil {
		return err
	}

	addKey, err := gResetSshPublicKeyDriver.GetAddPublicKey()
	if err != nil {
		return err
	}

	if addKey == "" && delKey == "" {
		return NoPublicKey
	}

	return gResetSshPublicKeyDriver.ResetPublicKey(delKey, addKey)
}
