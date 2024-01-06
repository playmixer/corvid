package yandex

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/playmixer/corvid/player"
)

func init() {

	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalln("Error loading .env file")
	}
}

func TestSpeach(t *testing.T) {
	ydx := New(os.Getenv("YANDEX_API_KEY"), os.Getenv("YANDEX_FOLDER_ID"))

	b, err := ydx.Speach("Текущее время 2 часа 30 минут").Post()
	if err != nil {
		t.Fatal(err)
	}

	player.PlayMp3FromBytes(context.Background(), b)
}
