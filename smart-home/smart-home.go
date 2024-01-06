package smarthome

import (
	"errors"
	"os"
	"strconv"
)

type TypeSHServiceName string

const (
	SHTuyaService TypeSHServiceName = "tuya"
)

type emptySmartHomeApi struct{}
type emptySmartHomeApiCommand struct{}

func (e *emptySmartHomeApi) GetDevice(device string) GetDeviceResponse {
	return GetDeviceResponse{}
}

func (e *emptySmartHomeApi) PostDevice(deviceId string, commands IApiCommand) {
	return
}

func (e *emptySmartHomeApi) NewCommand() IApiCommand {
	return &emptySmartHomeApiCommand{}
}

func (ec *emptySmartHomeApiCommand) Add(code string, value string) IApiCommand {
	return &emptySmartHomeApiCommand{}
}

type GetDeviceResponse struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
	T       int64       `json:"t"`
}

type PostDeviceCmdResponse struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Success bool   `json:"success"`
	Result  bool   `json:"result"`
	T       int64  `json:"t"`
}

type IApiCommand interface {
	Add(code string, value string) IApiCommand
}

type IApi interface {
	GetDevice(device string) GetDeviceResponse
	PostDevice(deviceId string, commands IApiCommand)
	NewCommand() IApiCommand
}

func FactoryNew(name TypeSHServiceName) (IApi, error) {
	smartHome, _ := strconv.ParseBool(os.Getenv("SMART_HOME"))
	if !smartHome {
		return &emptySmartHomeApi{}, nil
	}
	if name == SHTuyaService {
		return NewTuyaApi(), nil
	}

	return nil, errors.New("Factory not found service: " + string(name))
}
