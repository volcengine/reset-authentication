package pkg

import (
	"bufio"
	"github.com/volcengine/reset-authentication/util"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func needGrade(v1, v2 string) bool {
	ss1 := strings.Split(v1, ".")
	ss2 := strings.Split(v2, ".")

	if len(ss1) != len(ss2) && len(ss1) != 3 {
		return false
	}

	for i := 0; i < len(ss1); i++ {
		v1, err1 := strconv.Atoi(ss1[i])
		v2, err2 := strconv.Atoi(ss2[i])
		if err1 != nil || err2 != nil {
			return false
		}

		if v2 > v1 {
			return true
		}
	}

	return false
}

func getLatestVersion() (string, error) {
	var url = fmt.Sprintf("http://%s/volcstack/latest/reset_authentication_latest_version", gValidDataSource)
	latest, err := doDataSourceRequest(url, "GET", "")

	if err != nil {
		return "", err
	}
	return string(latest), nil
}

func getLatestDownloadUrl() (string, error) {
	var url = fmt.Sprintf("http://%s/volcstack/latest/reset_authentication_latest_url", gValidDataSource)
	downloadUrl, err := doDataSourceRequest(url, "GET", "")

	if err != nil {
		return "", err
	}
	return string(downloadUrl), nil
}

func download(url string, target io.Writer) error {

	client := http.DefaultClient
	client.Timeout = time.Second * 60

	resp, err := client.Get(url)
	if err != nil {
		return err
	}

	raw := resp.Body
	defer func() {
		_ = raw.Close()
	}()

	reader := bufio.NewReader(raw)
	_, err = io.Copy(target, reader)
	if err != nil {
		return err
	}

	return nil
}

func getExecPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}

	fileAbs, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}

	return fileAbs, nil
}

func Upgrade(version string) (string, bool) {
	if !util.GetConfig().AutoUpgrade {
		util.Info("Deny upgrade reset-authentication")
		return "", false
	}

	latestVersion, err := getLatestVersion()
	if err != nil {
		util.Error("Get latest reset-authentication version failed. Error:", err)
		return "", false
	}

	util.Info("Get latest reset-authentication version:", latestVersion, "current version", version)
	if latestVersion == "" {
		return "", false
	}

	if !needGrade(version, latestVersion) {
		util.Info("No need update reset-authentication")
		return "", false
	}

	downloadUrl, err := getLatestDownloadUrl()
	if err != nil {
		util.Error("Get latest reset-authentication download url failed. Error:", err)
		return "", false
	}

	path, err := getExecPath()
	if err != nil {
		util.Error("Get reset-authentication exec path failed. Error:", err)
		return "", false
	}

	var targetLatestPath = ""
	if runtime.GOOS != "windows" {
		targetLatestPath = filepath.Join(filepath.Dir(path), "reset-authentication_latest")
	} else {
		targetLatestPath = filepath.Join(filepath.Dir(path), "reset-authentication_latest.exe")
	}

	file, err := os.OpenFile(targetLatestPath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		util.Error("Create file failed. Error:", err)
		return "", false
	}

	if err = download(downloadUrl, file); err != nil {
		util.Error("Download latest reset-authentication failed. Error:", err)
		_ = file.Close()
		return "", false
	}

	_ = file.Close()

	util.Info("Download latest reset-authentication success")
	// Check that the updated software is available
	checkProgram := exec.Command(targetLatestPath, "--version")
	newV, err := checkProgram.Output()
	if err != nil {
		util.Error("Latest reset-authentication program start failed. Error:", err)
		return "", false
	}
	if string(newV) != latestVersion+"\n" {
		util.Error("Latest reset-authentication program start failed. Error:", err)
		return "", false
	}

	if runtime.GOOS != "windows" {
		if err = os.Rename(targetLatestPath, path); err != nil {
			util.Error("Rename "+targetLatestPath+" to "+path+" failed. Error:", err)
			_ = os.Remove(targetLatestPath)
			return "", false
		}

		targetLatestPath = path
	}

	util.Info("Upgrade latest reset-authentication success")
	return targetLatestPath, true
}
