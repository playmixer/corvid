package yandex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const ()

type RecognizeRequest struct {
	Lang            Lang   `json:"lang"`
	Topic           string `json:"topic"`
	ProfanityFilter bool   `json:"profanityFilter"`
	RawResults      bool   `json:"rawResults"`
	Format          Format `json:"format"`
	SampleRateHertz int    `json:"sampleRateHertz"`
	FolderId        string `json:"folderId"`
}

type RecognizeResponse struct {
	Result string `json:"result"`
}

func (y *YDX) RecognizeByte(data []byte) (RecognizeResponse, error) {
	result := RecognizeResponse{}
	if err := y.UpdIamToken(); err != nil {
		return result, err
	}

	// params := RecognizeRequest{
	// 	Lang:            LngRU,
	// 	Format:          FormatLPCM,
	// 	SampleRateHertz: 16000,
	// 	FolderId:        y.FolderId,
	// }

	params := map[string]interface{}{
		"lang":            LngRU,
		"form":            FormatLPCM,
		"sampleRateHertz": 16000,
		"folderId":        y.FolderId,
	}
	var queryParams []string
	for k := range params {
		queryParams = append(queryParams, fmt.Sprintf("%s=%s", k, params[k]))
	}
	url := URIRecognizeV1 + "?" + strings.Join(queryParams, "&")
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return result, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+y.IamToken)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}

	return result, err
}
