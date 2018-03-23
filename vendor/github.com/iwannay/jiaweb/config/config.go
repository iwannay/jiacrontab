package config

import (
	"errors"
	"io/ioutil"

	"github.com/iwannay/jiaweb/utils/file"
)

type (
	Config struct {
		App      *AppNode
		Server   *ServerNode
		Template *TemplateNode
		Jwt      *JwtNode
	}
	AppNode struct {
		Version     string
		Domain      string
		RunMode     string
		EnablePProf bool
		PProfPort   int
		LogPath     string
		EnableLog   bool
	}

	ServerNode struct {
		EnableListDir           bool
		EnableDetailRequestData bool
		EnableAutoHEAD          bool
		EnableIgnoreFavicon     bool
	}

	TemplateNode struct {
		TplDir string
		TplExt string
	}

	JwtNode struct {
		Domain       string
		Expire       int64
		Name         string
		EnableJwt    bool
		CookieMaxAge int64
		SignKey      string
	}
)

const (
	ConfigTypeJson = "json"
	ConfigTypeXml  = "xml"
)

func New() *Config {
	return &Config{
		App: &AppNode{
			EnableLog: false,
		},
		Server:   &ServerNode{},
		Template: &TemplateNode{},
	}
}

func InitConfig(configFile string, configType string) (config *Config, err error) {

	realPath := configFile
	if !file.Exist(configFile) {
		realPath = file.GetCurrentDirectory() + "/" + configFile
		if !file.Exist(realPath) {
			realPath = file.GetCurrentDirectory() + "/config/" + configFile

			if !file.Exist(realPath) {
				return nil, errors.New("no exists config file " + configFile)
			}
		}
	}

	if configType == ConfigTypeJson {
		config, err = initConfig(realPath, configType, fromJson)
	} else {
		return nil, errors.New("config not support xml type file")
	}

	if err != nil {
		return config, err
	}

	return config, nil

}

func initConfig(configFile string, configType string, f func([]byte, interface{}) error) (*Config, error) {
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, errors.New("jiaweb:config:initconfig " + err.Error())
	}

	var config *Config
	err = f(content, &config)
	if err != nil {
		return nil, errors.New("jiaweb:config:initconfig " + err.Error())
	}

	return config, nil

}
