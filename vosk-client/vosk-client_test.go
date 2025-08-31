package voskclient

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestPostConfigure(t *testing.T) {
	c := New()
	err := c.PostConfigure()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRecognize(t *testing.T) {
	c := New()
	c.Host = "192.168.0.2"
	err := c.PostConfigure()
	if err != nil {
		t.Fatal(err)
	}
	f, _ := os.Open("./test.wav")
	buf, _ := io.ReadAll(f)
	text, _ := c.Recognize(buf)
	fmt.Println(text)
}
