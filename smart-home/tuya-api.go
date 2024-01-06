package smarthome

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/tuya/tuya-connector-go/connector"
	"github.com/tuya/tuya-connector-go/connector/env"
)

type TuyaBodyCommand struct {
	Code  string `json:"code"`
	Value bool   `json:"value"`
}

type TuyaBody struct {
	Commands []TuyaBodyCommand `json:"commands"`
}

type TuyaAPI struct{}

func NewTuyaApi() *TuyaAPI {
	debug, _ := strconv.ParseBool(os.Getenv("TUYA_DEBUG"))
	connector.InitWithOptions(
		env.WithMsgHost(os.Getenv("TUYA_MSG_HOST")),
		env.WithApiHost(os.Getenv("TUYA_API_HOST")),
		env.WithAccessID(os.Getenv("TUYA_ACCESS_ID")),
		env.WithAccessKey(os.Getenv("TUYA_ACCESS_KEY")),
		env.WithAppName(os.Getenv("TUYA_APP_NAME")),
		env.WithDebugMode(debug),
	)

	return &TuyaAPI{}
}

func (t *TuyaAPI) GetDevice(deviceId string) GetDeviceResponse {
	return GetDeviceResponse{}
}

func (t *TuyaAPI) PostDevice(deviceId string, commands IApiCommand) {
	resp := &PostDeviceCmdResponse{}
	reqBody := commands

	body, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalln("Switch device failed:", err)
	}
	err = connector.MakePostRequest(
		context.Background(),
		connector.WithAPIUri(fmt.Sprintf("/v1.0/iot-03/devices/%s/commands", deviceId)),
		connector.WithPayload(body),
		connector.WithResp(resp),
	)
	if err != nil {
		log.Fatalln("Post body to switcher failed:", err)
	}
}

// return *TuyaBody
func (t *TuyaAPI) NewCommand() IApiCommand {
	return &TuyaBody{}
}

// *TuyaBody
func (tb *TuyaBody) Add(code string, value string) IApiCommand {
	tb.Commands = append(tb.Commands, TuyaBodyCommand{
		Code:  code,
		Value: value == "true",
	})

	return tb
}
