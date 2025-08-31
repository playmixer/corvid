package smarty

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	v "github.com/itchyny/volume-go"
	smarthome "github.com/playmixer/corvid/v1/smart-home"
	"go.uber.org/zap"
)

type TypeTool string

const (
	ToolWeather        TypeTool = "weather.current"
	ToolSysVolume      TypeTool = "system.volume"
	ToolSysVolumeParam TypeTool = "system.volume.param"
	ToolSysTime        TypeTool = "system.time"
	ToolTuyaSwitch     TypeTool = "smarthome.tuya.switch"
)

type CommandTool struct {
	f    func(ctx context.Context, a *Assiser, path string, args []string, ct *CommandTool) error
	temp string
}

func (ct *CommandTool) Run(ctx context.Context, a *Assiser, path string, args []string) error {
	return ct.f(ctx, a, path, args, ct)
}

var Tools = map[TypeTool]CommandTool{
	ToolWeather: {f: WeatherCurrent, temp: "Текущая температура {temp}, {descriptions}"},

	ToolSysVolume:      {f: SystemVolume, temp: ""},
	ToolSysVolumeParam: {f: SystemSetVolume, temp: ""},
	ToolSysTime:        {f: SystemTime, temp: "Текущее время {time}"},

	ToolTuyaSwitch: {f: SmartTuyaSwitch, temp: ""},
}

/**
* создаем команду для запуска tools
**/
func (a *Assiser) newCommandTool(pathFile string, args []string) CommandFunc {
	if _, ok := Tools[TypeTool(pathFile)]; ok {
		return func(ctx context.Context, a *Assiser) {
			if v, ok := Tools[TypeTool(pathFile)]; ok {
				err := v.Run(a.ctx, a, pathFile, args)
				if err != nil {
					a.log.Error("error", zap.Error(err))
					a.VoiceError(ctx, "Ошибка выполнения команды")
				}
			}
		}
	}

	return nil
}

type WeatherResult struct {
	Main struct {
		Temp     float32 `json:"temp"`
		TempMin  float32 `json:"temp_min"`
		TempMax  float32 `json:"temp_max"`
		FeelLike float32 `json:"feel_like"`
		Pressure int     `json:"pressure"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
	Wind struct {
		Speed float32 `json:"speed"`
	} `json:"wind"`
	Cloud struct {
		All int `json:"all"`
	} `json:"cloud"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
}

func WeatherCurrent(ctx context.Context, a *Assiser, path string, args []string, ct *CommandTool) error {
	apiKey := Getenv("OPENWEATHER_API_KEY", "")
	if apiKey == "" {
		return errors.New("not found openweather api key")
	}
	api := "https://api.openweathermap.org/data/2.5/weather"

	link, err := url.Parse(api)
	if err != nil {
		return err
	}
	queryParams := MarshalArgs(args)
	values := link.Query()
	for k, v := range queryParams {
		values.Add(k, v)
	}
	values.Add("appid", apiKey)
	// values.Add("units", "metric")

	link.RawQuery = values.Encode()
	resp, err := http.Get(link.String())
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("openweather return status: " + resp.Status + "\n\t" + link.String())
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	a.log.Debug("body", zap.ByteString("bytes", body))
	jBody := WeatherResult{}
	err = json.Unmarshal(body, &jBody)
	if err != nil {
		return err
	}
	var description = make([]string, 0)
	for i := range jBody.Weather {
		description = append(description, jBody.Weather[i].Description)
	}
	txt := strings.ReplaceAll(ct.temp, "{temp}", strconv.Itoa(int(jBody.Main.Temp)))
	txt = strings.ReplaceAll(txt, "{descriptions}", strings.Join(description, ","))
	err = a.Voice(ctx, txt)
	// err = a.Voice(fmt.Sprintf(`Текущая температура %s градусов, %s`, strconv.Itoa(int(jBody.Main.Temp)), strings.Join(description, ",")))
	if err != nil {
		return err
	}

	return nil
}

func SmartTuyaSwitch(ctx context.Context, a *Assiser, path string, args []string, ct *CommandTool) error {

	// API smart home
	sHome, err := smarthome.FactoryNew(smarthome.SHTuyaService)
	if err != nil {
		a.log.Error("error", zap.Error(err))
	}
	params := MarshalArgs(args)

	deviceid, ok := params["deviceid"]
	if !ok {
		return errors.New("tuya not found device id")
	}
	arrCodes, ok := params["code"]
	if !ok {
		return errors.New("tuya not found device code")
	}
	arrValues, ok := params["value"]
	if !ok {
		return errors.New("tuya not found values")
	}
	codes := strings.Split(arrCodes, "|")
	values := strings.Split(arrValues, "|")

	if len(codes) != len(values) {
		return errors.New("tuya: number of attributes does not match")
	}
	switcher := sHome.NewCommand()
	for i := range codes {
		switcher = switcher.Add(codes[i], values[i])
	}
	sHome.PostDevice(deviceid, switcher)

	return nil
}

func SystemTime(ctx context.Context, a *Assiser, path string, args []string, ct *CommandTool) error {
	if ct.temp != "" {
		txt := strings.ReplaceAll(ct.temp, "{time}", time.Now().Format("15 04"))
		a.Print(txt)
		err := a.Voice(ctx, txt)
		if err != nil {
			return err
		}
	}

	return nil
}

func SystemVolume(ctx context.Context, a *Assiser, path string, args []string, ct *CommandTool) error {
	params := MarshalArgs(args)

	curVol, err := v.GetVolume()
	if err != nil {
		log.Println(err.Error())
		return err
	}

	if up, ok := params["up"]; ok {
		vol, err := strconv.Atoi(up)
		if err != nil {
			log.Println(err.Error())
			return err
		}
		err = v.SetVolume(Min(curVol+vol, 100))
		if err != nil {
			log.Printf("set volume failed: %v", err)
			return err
		}
	}

	if down, ok := params["down"]; ok {
		vol, err := strconv.Atoi(down)
		if err != nil {
			log.Println(err.Error())
			return err
		}
		err = v.SetVolume(Max(curVol-vol, 0))
		if err != nil {
			log.Printf("set volume failed: %v", err)
			return err
		}
	}

	if volume, ok := params["volume"]; ok {
		vol, err := strconv.Atoi(volume)
		if err != nil {
			log.Println(err.Error())
			return err
		}
		err = v.SetVolume(vol)
		if err != nil {
			log.Printf("set volume failed: %v", err)
			return err
		}
	}

	return nil
}

func SystemSetVolume(ctx context.Context, a *Assiser, path string, args []string, ct *CommandTool) error {
	volume, ok := a.voice.Params["volume"]
	if !ok {
		return errors.New("param volume not found in command")
	}
	vol, err := strconv.Atoi(volume)
	if err != nil {
		return errors.New("param volume not is integer")
	}

	err = v.SetVolume(vol)
	if err != nil {
		log.Printf("set volume failed: %s", err)
		return err
	}

	return nil
}
