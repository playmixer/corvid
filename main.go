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
	recognizer.Host = "192.168.0.2"
	recognizer.Port = "2700"
	recognizer.SetLogger(log)

	assistent := smarty.New(ctx)
	assistent.SetRecognizeCommand(recognizer)
	assistent.SetRecognizeName(recognizer)
	assistent.SetLogger(log)
	assistent.SetConfig(smarty.Config{
		Names:           []string{"альфа", "бета", "бэта"},
		ListenLongTime:  time.Second / 2,
		LenWavBuf:       40,
		MaxEmptyMessage: 40,
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
