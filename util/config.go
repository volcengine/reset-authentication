package util

import (
	"flag"
	"runtime"
)

type Config struct {
	DataSource struct {
		Addresses string
		Timeout   int
		Retries   int
		Interval  int
	}
	LogPath              string
	EnableResetPassword  bool
	EnableResetPublicKey bool
	AutoUpgrade          bool
}

var gConfig Config

func GetConfig() *Config {
	return &gConfig
}

var Version bool

func setConfigDefault() {
	flag.StringVar(&(gConfig.DataSource.Addresses), "datasource-address", "100.96.0.96 169.254.169.254", "Specify data source addresses")
	flag.IntVar(&(gConfig.DataSource.Timeout), "datasource-timeout", 5, "Specify data source timeout")
	flag.IntVar(&(gConfig.DataSource.Retries), "datasource-retries", 10, "Specify data source retries")
	flag.IntVar(&(gConfig.DataSource.Interval), "datasource-interval", 5, "Specify data source interval")
	if runtime.GOOS == "windows" {
		flag.StringVar(&(gConfig.LogPath), "logpath", "C:\\Program Files\\Reset Authentication\\log", "Specify log path")
	} else {
		flag.StringVar(&(gConfig.LogPath), "logpath", "/var/log/volcstack/", "Specify log path")
		flag.BoolVar(&(gConfig.EnableResetPassword), "enable-reset-password", true, "Specify enable reset password")
		flag.BoolVar(&(gConfig.EnableResetPublicKey), "enable-reset-pubkey", true, "Specify enable reset ssh authorized keys")
	}
	flag.BoolVar(&(gConfig.AutoUpgrade), "auto-update", true, "Specify enable upgrade reset-authentication")
}

func setVersionFlag() {
	flag.BoolVar(&Version, "version", false, "Show reset-authentication version")
}

func init() {
	setConfigDefault()
	setVersionFlag()

	flag.Parse()
}
