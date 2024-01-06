package smarty

import (
	"context"
	"log"
	"os"
	"strings"

	v "github.com/itchyny/volume-go"
)

// объеденяем каналы
func MergeChans(ctx context.Context, c1 chan []byte, c2 chan []byte) chan []byte {
	result := make(chan []byte, 1)
	go func() {
	waitChan2:
		for {
			select {
			case <-ctx.Done():
				break waitChan2
			case r := <-c1:
				result <- r
			case r := <-c2:
				result <- r
			}

		}
	}()
	return result
}

func Getenv(key string, def string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}

	return def
}

func MarshalArgs(args []string) map[string]string {
	var result = make(map[string]string)
	for _, arg := range args {
		_r := strings.Split(arg, "=")
		if len(_r) == 2 {
			result[_r[0]] = _r[1]
		}
	}
	return result
}

func GetVolume() int {
	vol, err := v.GetVolume()
	if err != nil {
		log.Fatalf("get volume failed: %+v", err)
	}
	return vol
}

func SetVolume(vol int) {
	err := v.SetVolume(vol)
	if err != nil {
		log.Fatalf("set volume failed: %+v", err)
	}
}

func Max(val1, val2 int) int {
	if val1 > val2 {
		return val1
	}
	return val2
}

func Min(val1, val2 int) int {
	if val1 < val2 {
		return val1
	}
	return val2
}
