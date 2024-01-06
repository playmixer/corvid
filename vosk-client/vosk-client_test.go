package voskclient

import (
	"testing"
)

func TestPostConfigure(t *testing.T) {
	c := New()
	err := c.PostConfigure()
	if err != nil {
		t.Fatal(err)
	}
}
