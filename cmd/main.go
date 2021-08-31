package main

import (
	"github.com/volcengine/reset-authentication/pkg"
	"github.com/volcengine/reset-authentication/util"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

var (
	version string
)

func checkValidDataSource() {
	for i := 0; i < util.GetConfig().DataSource.Retries; i++ {
		err := pkg.GetValidDataSource()
		if err == nil {
			break
		}

		if i == util.GetConfig().DataSource.Retries-1 {
			util.Error("No valid data source.")
			os.Exit(-1)
		}

		time.Sleep(time.Second * time.Duration(util.GetConfig().DataSource.Interval))
	}
}

func pushDataSourceVersion() {
	// push version information to confirm whether a function is supported
	if version != "" {
		err := pkg.PushVersion(version)
		if err != nil {
			util.Error("Push version failed, error:", err)
		}
	} else {
		util.Error("The version is nil.")
		os.Exit(-1)
	}
}

func autoUpgrade() {
	targetExecFile, b := pkg.Upgrade(version)
	if !b {
		return
	}

	util.Info("Start the latest reset-authentication program")

	var newProgram *exec.Cmd = nil
	if len(os.Args) >= 2 {
		newProgram = exec.Command(targetExecFile, os.Args[1:]...)
	} else {
		newProgram = exec.Command(targetExecFile)
	}

	_, err := newProgram.Output()
	if err != nil {
		util.Error("start new reset-authentication program failed. Error:", err)
	} else {
		// use latest reset-authentication to reset password and public-key.
		os.Exit(0)
	}
}

func resetPassword() {
	err := pkg.ResetPassword()
	if err != nil {
		if err == pkg.NoPassword {
			util.Info("No password needs to be changed.")
		} else {
			util.Error("Reset password failed. Error:", err)
		}
	} else {
		util.Info("Reset password success.")
	}
}

func resetPublicKey() {
	if runtime.GOOS != "windows" {
		err := pkg.ResetSshPublicKey()
		if err != nil {
			if err == pkg.NoPublicKey || err == pkg.NoResetPubkeyDriver {
				util.Info(err)
			} else {
				util.Error("Reset pubkey failed. Error:", err)
			}
		} else {
			util.Info("Reset pubkey success.")
		}
	}
}

func main() {

	if util.Version {
		if version != "" {
			_, _ = fmt.Fprintln(os.Stdout, version)
			os.Exit(0)
		} else {
			_, _ = fmt.Fprintln(os.Stderr, "Get version failed")
			os.Exit(-1)
		}
	}

	util.Info(util.GetConfig())

	checkValidDataSource()

	autoUpgrade()

	pushDataSourceVersion()

	resetPassword()

	resetPublicKey()
}
