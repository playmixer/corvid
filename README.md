# Corvid - библиотке для создания голосовго ассистента

## Package
```
go get github.com/playmixer/corvid
```


## DEMO
[demo](./main.go)

### Создаём Dockerfile для запуска сервера распознования речи
```docker
FROM alphacep/kaldi-vosk-server:latest

ENV MODEL vosk-model-ru-0.22
RUN mkdir /opt/vosk-model \
   && cd /opt/vosk-model \
   && wget -q https://alphacephei.com/vosk/models/${MODEL}.zip \
   && unzip ${MODEL}.zip \
   && mv ${MODEL} model \
   && rm -rf model/extra \
   && rm -rf ${MODEL}.zip

EXPOSE 2700
WORKDIR /opt/vosk-server/websocket
CMD [ "python3", "./asr_server.py", "/opt/vosk-model/model" ]
```

### Запускаем vosk-server
```
docker build -t vosk-server . && docker run -p 2700:2700 -it vosk-server
```

### Создаем голосового помощника
```
go get github.com/playmixer/corvid
```
```golang
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/playmixer/corvid/logger"
	"github.com/playmixer/corvid/smarty"
	voskclient "github.com/playmixer/corvid/vosk-client"
)

func main() {

	ctx := context.TODO()

	log := logger.New("app")
	log.LogLevel = logger.INFO

	recognizer := voskclient.New()
	recognizer.SetLogger(log)

	assistent := smarty.New(ctx)
	assistent.SetRecognizeCommand(recognizer)
	assistent.SetRecognizeName(recognizer)
	assistent.SetLogger(log)
	assistent.SetConfig(smarty.Config{
		Names:         []string{"альфа", "бета", "бэта"},
		ListenTimeout: time.Second * 1,
	})

	// Голосовые команды
	assistent.AddCommand([]string{"который час"}, func(ctx context.Context, a *smarty.Assiser) {
		txt := fmt.Sprint("Текущее время:", time.Now().Format(time.TimeOnly))
		a.Print(txt)
	})

	log.INFO("Starting App")
	assistent.Start()

	log.INFO("Stop App")
}

```
### Запускаем
```golang
go run .
```
*Произнести команду в микрофон **"альфа который час"***

### Настроить голос асистенту через Yandex API
```golang
	...
	"github.com/playmixer/corvid/yandex"
	"github.com/playmixer/corvid"
	...
	assistent.SetTTS(func(text string) error {
		ydx := yandex.New(os.Getenv("YANDEX_API_KEY"), os.Getenv("YANDEX_FOLDER_ID"))

		req := ydx.Speach(text)
		b, err := req.Post()
		if err != nil {
			return err
		}

		player.PlayMp3FromBytes(b)
		return nil
	})
	...

```

## Добавить команду из набора
- добавить API ключ *openweathermap.org* в **.env**

- получить ключь можно на https://home.openweathermap.org/api_keys
```env
OPENWEATHER_API_KEY={api key}
```
- подробное описание параметров на https://openweathermap.org/current
#### Пример:
```golang
	...
	// Текущая погода
	assistent.AddGenCommand(smarty.ObjectCommand{
			Type:     smarty.TypeCommand("tool"),
			Path:     "weather.current",
			Args:     []string{
				"lat=54.974897",
				"lon=73.4777",
				"units=metric",
				"lang=ru"
			},
			Commands: []string{
				"какая сейчас погода",
			},
		})
	// или альтернативный вариант	
	assistent.AddGenCommand(smarty.ObjectCommand{
			Type:     smarty.TypeCommand("tool"),
			Path:     "weather.current",
			Args:     []string{
				"q=Омск",
				"units=metric",
				"lang=ru"
			},
			Commands: []string{
				"какая сейчас погода",
			},
		})
	...

```

## Добавить умный выключатель Tuya
```env
TUYA_MSG_HOST=pulsar+ssl://mqe.tuyaeu.com:7285/
TUYA_API_HOST=https://openapi.tuyaeu.com
TUYA_ACCESS_ID={access_id}
TUYA_ACCESS_KEY={access_key}
TUYA_APP_NAME={app_name}
TUYA_DEBUG=0
```

```golang
	...
	assistent.AddGenCommand(smarty.ObjectCommand{
			Type:     smarty.TypeCommand("tool"),
			Path:     "smarthome.tuya.switch",
			Args:     []string{
				"deviceid=787166238ccb3a259625",
				"code=switch_1",
				"value=true",
			},
			Commands: []string{
				"включи свет",
			},
		})
	...
	
```
