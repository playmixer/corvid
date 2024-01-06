package voskclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

var (
	log logger = _l{}
)

func getenv(key string, d string) string {
	v := os.Getenv(key)
	if len(v) == 0 {
		return d
	}
	return v
}

type logger interface {
	ERROR(v ...string)
	INFO(v ...string)
	DEBUG(v ...string)
}

type _l struct{}

func (l _l) ERROR(v ...string) {
	fmt.Println(v)
}
func (l _l) INFO(v ...string) {
	fmt.Println(v)
}
func (l _l) DEBUG(v ...string) {
	fmt.Println(v)
}

type Client struct {
	Host       string
	Port       string
	buffsize   int
	SampleRate int
	socket     *websocket.Conn
}

func New() *Client {
	clt := &Client{
		Host:       getenv("VOSK_HOST", "localhost"),
		Port:       getenv("VOSK_PORT", "2700"),
		buffsize:   8000,
		SampleRate: 16000,
	}

	return clt
}

type Message struct {
	Result []struct {
		Conf  float64
		End   float64
		Start float64
		Word  string
	}
	Text string
}

var m Message

func (c *Client) SetLogger(l logger) {
	log = l
}

func (c *Client) PostConfigure() error {
	u := url.URL{Scheme: "ws", Host: c.Host + ":" + c.Port, Path: ""}
	log.DEBUG("connecting to ", u.String())

	// Opening websocket connection
	soc, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.ERROR(err.Error())
		return err
	}
	c.socket = soc
	defer c.socket.Close()

	err = soc.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("{\"config\" : { \"sample_rate\" : %v } }", c.SampleRate)))
	if err != nil {
		log.ERROR(err.Error())
		return err
	}

	// err = c.socket.WriteMessage(websocket.TextMessage, []byte("{\"eof\" : 1}"))
	// if err != nil {
	// 	log.ERROR(err.Error())
	// 	return err
	// }
	// Closing websocket connection
	err = c.socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.ERROR(err.Error())
		return err
	}

	return nil
}

func (c *Client) Recognize(bufWav []byte) (string, error) {
	u := url.URL{Scheme: "ws", Host: c.Host + ":" + c.Port, Path: ""}
	log.DEBUG("connecting to ", u.String())

	// Opening websocket connection
	soc, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.ERROR(err.Error())
		return "", err
	}
	c.socket = soc
	defer c.socket.Close()

	err = soc.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("{\"config\" : { \"sample_rate\" : %v } }", c.SampleRate)))
	if err != nil {
		log.ERROR(err.Error())
		return "", err
	}

	f := bytes.NewReader(bufWav)
	for {
		buf := make([]byte, c.buffsize)

		dat, err := f.Read(buf)

		if dat == 0 && err == io.EOF {
			err = c.socket.WriteMessage(websocket.TextMessage, []byte("{\"eof\" : 1}"))
			if err != nil {
				log.ERROR(err.Error())
				return "", err
			}
			break
		}
		if err != nil {
			log.ERROR(err.Error())
			return "", err
		}

		err = c.socket.WriteMessage(websocket.BinaryMessage, buf)
		if err != nil {
			log.ERROR(err.Error())
			return "", err
		}

		// Read message from server
		_, _, err = c.socket.ReadMessage()
		if err != nil {
			log.ERROR(err.Error())
			return "", err
		}
	}

	// Read final message from server
	_, msg, err := c.socket.ReadMessage()
	if err != nil {
		log.ERROR(err.Error())
		return "", err
	}

	// Closing websocket connection
	c.socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	// Unmarshalling received message
	err = json.Unmarshal(msg, &m)
	if err != nil {
		log.ERROR(err.Error())
		return "", err
	}
	return m.Text, nil
}
