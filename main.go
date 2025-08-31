package main

import (
	"context"
	"fmt"
	"log"
	"time"

	logger "github.com/playmixer/corvid/logger/v2"

	"github.com/playmixer/corvid/v1/smarty"
	voskclient "github.com/playmixer/corvid/v1/vosk-client"
)

func main() {
	ctx := context.TODO()

	lgr, err := logger.New()
	if err != nil {
		log.Fatal(err)
	}

	recognizer := voskclient.New()
	recognizer.Host = "192.168.0.2"
	recognizer.Port = "2700"

	assistent := smarty.New(ctx)
	assistent.SetRecognizeCommand(recognizer)
	assistent.SetRecognizeName(recognizer)
	assistent.SetLogger(lgr)
	assistent.SetConfig(smarty.Config{
		Names:           []string{"альфа"},
		ListenLongTime:  time.Second / 2,
		LenWavBuf:       40,
		MaxEmptyMessage: 40,
	})

	// Голосовые команды
	assistent.AddCommand([]string{"который час"}, func(ctx context.Context, a *smarty.Assiser) {
		txt := fmt.Sprint("Текущее время:", time.Now().Format(time.TimeOnly))
		a.Print(txt)
	})

	log.Println("Starting App")
	assistent.Start()

	log.Println("Stop App")
}
