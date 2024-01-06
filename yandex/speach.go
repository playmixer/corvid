package yandex

import (
	"encoding/json"
	"errors"
)

type SpeachResponse struct {
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

type SpeachRequest struct {
	Text string `json:"text"`
	// SSML            string  `json:"ssml"`
	Lang            Lang    `json:"lang"`
	Voice           Voice   `json:"voice"`
	Emotion         Emotion `json:"emotion"`
	Speed           string  `json:"speed"`
	Format          Format  `json:"format"`
	SampleRateHertz string  `json:"sampleRateHertz"`
	FolderId        string  `json:"folderId"`
	y               *YDX    `json:"_"`
}

func (y *YDX) Speach(text string) *SpeachRequest {
	query := SpeachRequest{
		Text:            text,
		Lang:            LngRU,
		Voice:           VoiceJane,
		Emotion:         EmotionNeutral,
		Speed:           "1.0",
		Format:          FormatMP3,
		SampleRateHertz: "16000",
		FolderId:        y.FolderId,
		y:               y,
	}

	return &query
}

func (r *SpeachRequest) Post() ([]byte, error) {
	if r.y.FolderId == "" {
		return nil, errors.New("Yandex speach request error: Invalid folderId")
	}
	if r.y.OAuthToken == "" {
		return nil, errors.New("Yandex speach request error: Invalid OAuthToken")
	}

	_b, err := json.Marshal(*r)
	if err != nil {
		return nil, err
	}
	params := make(map[string]string)
	err = json.Unmarshal(_b, &params)
	if err != nil {
		return nil, err
	}
	res, err := r.y.postRequest(URISpeachV1, map[string]string{}, []byte{}, params)
	if err != nil {
		return nil, err
	}

	body := SpeachResponse{}
	json.Unmarshal(res, &body)
	if body.ErrorCode != "" {
		return nil, errors.New(body.ErrorCode + ": " + body.ErrorMessage)
	}

	return res, nil
}
