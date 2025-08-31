package player

import (
	"bytes"
	"context"
	"io"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
)

func PlayWavFromBytes(ctx context.Context, b []byte) error {

	streamer, format, err := wav.Decode(bytes.NewReader(b))
	defer streamer.Close()
	if err != nil {
		return err
	}
	return playstream(ctx, streamer, format)
}

func PlayOggFromBytes(ctx context.Context, b []byte) error {
	streamer, format, err := vorbis.Decode(io.NopCloser(bytes.NewReader(b)))
	defer streamer.Close()
	if err != nil {
		return err
	}
	return playstream(ctx, streamer, format)
}

func PlayMp3FromBytes(ctx context.Context, b []byte) error {
	streamer, format, err := mp3.Decode(io.NopCloser(bytes.NewReader(b)))
	defer streamer.Close()
	if err != nil {
		return err
	}
	return playstream(ctx, streamer, format)
}

func PlayWavFromFile(ctx context.Context, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	streamer, format, err := wav.Decode(f)
	defer streamer.Close()
	if err != nil {
		return err
	}
	return playstream(ctx, streamer, format)
}

func playstream(ctx context.Context, streamer beep.StreamSeekCloser, format beep.Format) error {
	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		return err
	}
	done := make(chan bool)
	defer close(done)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	select {
	case <-done:
	case <-ctx.Done():
		speaker.Close()
	}

	return nil
}
