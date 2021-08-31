package pkg

import (
	"bufio"
	"github.com/volcengine/reset-authentication/util"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var gValidDataSource string

func doDataSourceRequest(url string, method string, request string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest(method, url, strings.NewReader(request))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Second*time.Duration(util.GetConfig().DataSource.Timeout))
	defer cancel()

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(out))
	}

	return out, nil
}

func getInstanceId(addr string) (string, error) {
	var url = fmt.Sprintf("http://%s/volcstack/latest/instance_id", addr)
	idBytes, err := doDataSourceRequest(url, "GET", "")
	if err != nil {
		return "", err
	}

	return string(idBytes), err
}

func GetValidDataSource() error {
	var addrList = strings.Split(util.GetConfig().DataSource.Addresses, " ")

	for _, addr := range addrList {
		if addr == "" {
			continue
		}

		_, err := getInstanceId(addr)
		if err == nil {
			gValidDataSource = addr
			return nil
		}

		util.Error("Check Datasource", addr, "failed. Error:", err)
	}

	return NoValidDataSource
}

func PushVersion(version string) error {
	var url = fmt.Sprintf("http://%s/volcstack/latest/reset_authentication_version", gValidDataSource)
	_, err := doDataSourceRequest(url, "POST", version)

	if err != nil {
		return err
	}
	return nil
}

func GetLatestVersion() (string, error) {
	var url = fmt.Sprintf("http://%s/volcstack/latest/reset_authentication_latest_version", gValidDataSource)
	latest, err := doDataSourceRequest(url, "GET", "")

	if err != nil {
		return "", err
	}
	return string(latest), nil
}

func GetLatestDownloadUrl() (string, error) {
	var url = fmt.Sprintf("http://%s/volcstack/latest/reset_authentication_latest_url", gValidDataSource)
	downloadUrl, err := doDataSourceRequest(url, "GET", "")

	if err != nil {
		return "", err
	}
	return string(downloadUrl), nil
}

func Download(url string, target io.Writer) error {

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
