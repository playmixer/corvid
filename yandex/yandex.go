package yandex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	URIOAuthToken  = "https://iam.api.cloud.yandex.net/iam/v1/tokens"
	URIRecognizeV1 = "https://stt.api.cloud.yandex.net/speech/v1/stt:recognize"
	URISpeachV1    = "https://tts.api.cloud.yandex.net/speech/v1/tts:synthesize"
)

type Lang string

const (
	LngAuto Lang = "auto"  //автоматическое распознавание языка
	LngDE   Lang = "de-DE" //немецкий
	LngUS   Lang = "en-US" //английский
	LngES   Lang = "es-ES" //испанский
	LngFI   Lang = "fi-FI" //финский
	LngFR   Lang = "fr-FR" //французский
	LngHE   Lang = "he-HE" //иврит
	LngIT   Lang = "it-IT" //итальянский
	LngKZ   Lang = "kk-KZ" //казахский
	LngNL   Lang = "nl-NL" //голландский
	LngPL   Lang = "pl-PL" //польский
	LngPT   Lang = "pt-PT" //португальский
	LngBR   Lang = "pt-BR" //бразильский португальский
	LngRU   Lang = "ru-RU" //русский язык (по умолчанию)
	LngSE   Lang = "sv-SE" //шведский
	LngTR   Lang = "tr-TR" //турецкий
	LngUZ   Lang = "uz-UZ" //узбекский (латиница)
)

type Voice string

const (
	VoiceAlena     Voice = "alena"     //(по умолчанию)	Ж	(по умолчанию) нейтральная — neutral радостная — good v1, v3
	VoiceFilipp    Voice = "filipp"    //	M	—	v1, v3 	Voiceermil = "ermil" //	M	(по умолчанию) нейтральный — neutral радостный — good	v1, v3
	VoiceJane      Voice = "jane"      //	Ж	(по умолчанию) нейтральная — neutral радостная — good 	раздраженная — evil	v1, v3
	VoiceMadirus   Voice = "madirus"   //	M	—	v1, v3 	Voiceomazh = "omazh" //	Ж	(по умолчанию) нейтральная — neutral раздраженная — evil	v1, v3
	VoiceZahar     Voice = "zahar"     //	M	(по умолчанию) нейтральный — neutral радостный — good	v1, v3
	VoiceDasha     Voice = "dasha"     //	Ж	(по умолчанию) нейтральная — neutral радостная — good 	дружелюбная — friendly	v3
	VoiceJulia     Voice = "julia"     //	Ж	(по умолчанию) нейтральная — neutral строгая — strict	v3
	VoiceLera      Voice = "lera"      //	Ж	(по умолчанию) нейтральная — neutral дружелюбная — friendly	v3
	VoiceMasha     Voice = "masha"     //	Ж	(по умолчанию) радостная — good строгая — strict дружелюбная — friendly	v3
	VoiceMarina    Voice = "marina"    //	Ж	(по умолчанию) нейтральная — neutral шепот — whisper дружелюбная — friendly	v3
	VoiceAlexander Voice = "alexander" //	M	(по умолчанию) нейтральный — neutral радостный — good	v3
	VoiceKirill    Voice = "kirill"    //	M	(по умолчанию) нейтральный — neutral строгий — strict радостный — good	v3
	VoiceAnton     Voice = "anton"     //	M	(по умолчанию) нейтральный — neutral радостный — good	v3
)

type Emotion string

const (
	EmotionModer    Emotion = "modern"
	EmotionClassic  Emotion = "classic"
	EmotionNeutral  Emotion = "neutral"
	EmotionGood     Emotion = "good"
	EmotionEvil     Emotion = "evil"
	EmotionFriendly Emotion = "friendly"
	EmotionStrict   Emotion = "strict"
	EmotionWisper   Emotion = "whisper"
)

type Format string

const (
	FormatLPCM Format = "lpcm"
	FormatOGG  Format = "oggopus"
	FormatMP3  Format = "mp3"
)

type YDX struct {
	FolderId        string
	OAuthToken      string
	IamToken        string
	IamTokenExpires time.Time
}

func New(OAuthToken, folderId string) *YDX {

	return &YDX{
		OAuthToken: OAuthToken,
		FolderId:   folderId,
	}
}

type IAMTokenRequest struct {
	YandexPassportOauthToken string `json:"yandexPassportOauthToken"`
}

type IAMTOkenRespons struct {
	IamToken  string `json:"iamToken"`
	ExpiresAt string `json:"expiresAt"`
}

func (y *YDX) getIAMToken() (IAMTOkenRespons, error) {
	sBody := IAMTOkenRespons{}

	data := IAMTokenRequest{
		YandexPassportOauthToken: y.OAuthToken,
	}
	dataByte, err := json.Marshal(data)
	if err != nil {
		return sBody, err
	}

	req, err := http.NewRequest(http.MethodPost, URIOAuthToken, bytes.NewBuffer(dataByte))
	if err != nil {
		return sBody, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return sBody, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return sBody, err
	}

	err = json.Unmarshal(body, &sBody)

	return sBody, err

}

func (y *YDX) UpdIamToken() error {
	fmt.Println("time duration", time.Until(y.IamTokenExpires), time.Hour)
	if y.IamToken != "" && time.Until(y.IamTokenExpires) > time.Hour {
		return nil
	}
	res, err := y.getIAMToken()
	if err != nil {
		return err
	}

	y.IamToken = res.IamToken
	y.IamTokenExpires, err = time.Parse("2006-01-02T15:04:05.999999999Z", res.ExpiresAt)
	if err != nil {
		return err
	}
	return nil
}

func (y *YDX) postRequest(uri string, headers map[string]string, body []byte, request map[string]string) ([]byte, error) {
	var result []byte

	if request != nil {
		baseUrl, err := url.Parse(uri)
		if err != nil {
			return nil, err
		}
		params := url.Values{}
		for k, v := range request {
			params.Add(k, v)
		}
		baseUrl.RawQuery = params.Encode()
		uri = baseUrl.String()
	}
	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
	if err != nil {
		return result, err
	}
	// req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	// req.Header.Set("Authorization", "Bearer "+y.IamToken)
	req.Header.Set("Authorization", "Api-Key "+y.OAuthToken)
	for _, k := range headers {
		req.Header.Set(k, headers[k])
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer res.Body.Close()

	result, err = io.ReadAll(res.Body)
	if err != nil {
		return result, err
	}

	return result, err
}
