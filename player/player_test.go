package player

import (
	"context"
	"io"
	"os"
	"testing"
	"time"
)

func TestPlayMp3FromBytesCancelContext(t *testing.T) {
	f, err := os.Open("./Текущая температура -17, небольшая облачность.mp3")
	if err != nil {
		t.Fatal(err)
	}

	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	PlayMp3FromBytes(ctx, b)
	time.Sleep(time.Second * 4)
}
