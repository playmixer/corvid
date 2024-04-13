package smarty_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/playmixer/corvid/smarty"
)

var (
	assist     *smarty.Assiser
	idsCommand map[string]string
)

type rcgnz struct{}

func (r *rcgnz) Recognize(bufWav []byte) (string, error) {
	return "тест", nil
}

func Init() {
	godotenv.Load()

	ctx := context.TODO()
	recognize := &rcgnz{}
	assist = smarty.New(ctx)
	assist.SetRecognizeCommand(recognize)
	assist.SetRecognizeName(recognize)

	idsCommand = make(map[string]string)

	idsCommand["test"] = assist.AddCommand([]string{"тест"}, func(ctx context.Context, a *smarty.Assiser) {})                                                          //0
	idsCommand["который час"] = assist.AddCommand([]string{"который час", "какое время", "сколько времени"}, func(ctx context.Context, a *smarty.Assiser) {})          //1
	idsCommand["включи свет в ванне"] = assist.AddCommand([]string{"включи свет в ванне", "включи в ванной свет"}, func(ctx context.Context, a *smarty.Assiser) {})    //2
	idsCommand["выключи свет в ванне"] = assist.AddCommand([]string{"выключи свет в ванне", "выключи в ванной свет"}, func(ctx context.Context, a *smarty.Assiser) {}) //3
	idsCommand["запусти браузер"] = assist.AddCommand([]string{"запусти браузер"}, func(ctx context.Context, a *smarty.Assiser) {})                                    //4
	idsCommand["включи стим"] = assist.AddCommand([]string{"включи стим"}, func(ctx context.Context, a *smarty.Assiser) {})                                            //5
	idsCommand["отключись"] = assist.AddCommand([]string{"отключись", "выключись"}, func(ctx context.Context, a *smarty.Assiser) {})                                   //6
}

func TestRotateCommand(t *testing.T) {
	Init()
	type testRotate struct {
		cmd string
		i   string
		p   int
	}

	cases := []testRotate{
		{"тсет", "", 0},                   //0
		{"тест", idsCommand["test"], 100}, //1
		{"скажи который час", idsCommand["который час"], 100},                      //2
		{"какое сейчас время", idsCommand["который час"], 100},                     //3
		{"сколько сейчас времени", idsCommand["который час"], 100},                 //4
		{"подскажи время", "", 0},                                                  //5
		{"включи", "", 0},                                                          //6
		{"включи свет в ванне пожалуйста", idsCommand["включи свет в ванне"], 100}, //7
		{"включи свет в ванной", idsCommand["включи свет в ванне"], 100},           //8
		{"выключить свет в ванной", "", 0},                                         //9
		{"выключи свет в ванной", idsCommand["выключи свет в ванне"], 100},         //10
	}
	for idx, c := range cases {
		if i, p := assist.RotateCommand(c.cmd); i != c.i || p != c.p {
			t.Fatalf("case#%v idx#%v percent#%v %s", idx, i, p, c.cmd)
		}
	}

}
func TestFoundCommandByToken(t *testing.T) {
	Init()
	type testRotate struct {
		cmd   string
		i     string
		p     int
		found bool
	}

	cases := []testRotate{
		{"тсет", idsCommand["test"], 75, true},                                           //0
		{"тест", idsCommand["test"], 100, true},                                          //1
		{"скажи который час", idsCommand["который час"], 100, true},                      //2
		{"какое сейчас время", idsCommand["который час"], 100, true},                     //3
		{"сколько сейчас времени", idsCommand["который час"], 100, true},                 //4
		{"подскажи время", idsCommand["который час"], 64, true},                          //5
		{"включи", idsCommand["включи свет в ванне"], 100, true},                         //6
		{"включи свет в ванне пожалуйста", idsCommand["включи свет в ванне"], 100, true}, //7
		{"включи свет в ванной", idsCommand["включи свет в ванне"], 100, true},           //8
		{"выключить свет в ванной", idsCommand["выключи свет в ванне"], 95, true},        //9
		{"выключи свет в ванной", idsCommand["выключи свет в ванне"], 100, true},         //10
		{"запусти стим", idsCommand["запусти браузер"], 74, true},                        //11
		{"отключись", idsCommand["отключись"], 100, true},                                //12
		{"включить", idsCommand["отключись"], 82, true},                                  //13
	}
	for idx, c := range cases {
		if i, p, found := assist.FoundCommandByToken(c.cmd); i != c.i || p != c.p || found != c.found {
			t.Fatalf("case#%v idx=%v percent=%v found=%v %s", idx, i, p, found, c.cmd)
		}
	}

}
func TestFoundCommandByDistance(t *testing.T) {
	Init()
	type testRotate struct {
		cmd   string
		i     string
		d     int
		found bool
	}

	cases := []testRotate{
		{"тсет", idsCommand["test"], 2, true},                                           //0
		{"тест", idsCommand["test"], 0, true},                                           //1
		{"скажи который час", idsCommand["который час"], 6, true},                       //2
		{"какое сейчас время", idsCommand["который час"], 7, true},                      //3
		{"сколько сейчас времени", idsCommand["который час"], 7, true},                  //4
		{"подскажи время", idsCommand["который час"], 7, true},                          //5
		{"включи", idsCommand["отключись"], 3, true},                                    //6
		{"включи свет в ванне пожалуйста", idsCommand["включи свет в ванне"], 11, true}, //7
		{"включи свет в ванной", idsCommand["включи свет в ванне"], 2, true},            //8
		{"выключить свет в ванной", idsCommand["выключи свет в ванне"], 4, true},        //9
		{"выключи свет в ванной", idsCommand["выключи свет в ванне"], 2, true},          //10
		{"запусти стим", idsCommand["включи стим"], 6, true},                            //11
		{"отключись", idsCommand["отключись"], 0, true},                                 //12
		{"включить", idsCommand["отключись"], 2, true},                                  //13
	}
	for idx, c := range cases {
		if i, d, found := assist.FoundCommandByDistance(c.cmd); i != c.i || d != c.d || found != c.found {
			t.Fatalf("case#%v idx=%v distance=%v found=%v %s", idx, i, d, found, c.cmd)
		}
	}

}
func TestFoundCommandByRatio(t *testing.T) {
	Init()
	type testRotate struct {
		cmd   string
		i     string
		r     int
		found bool
	}

	cases := []testRotate{
		{"тсет", idsCommand["test"], 75, true},                                          //0
		{"тест", idsCommand["test"], 100, true},                                         //1
		{"скажи который час", idsCommand["который час"], 79, true},                      //2
		{"какое сейчас время", idsCommand["который час"], 76, true},                     //3
		{"сколько сейчас времени", idsCommand["который час"], 81, true},                 //4
		{"подскажи время", idsCommand["который час"], 64, true},                         //5
		{"включи", idsCommand["отключись"], 80, true},                                   //6
		{"включи свет в ванне пожалуйста", idsCommand["включи свет в ванне"], 78, true}, //7
		{"включи свет в ванной", idsCommand["включи свет в ванне"], 92, true},           //8
		{"выключить свет в ванной", idsCommand["выключи свет в ванне"], 88, true},       //9
		{"выключи свет в ванной", idsCommand["выключи свет в ванне"], 93, true},         //10
		{"запусти стим", idsCommand["запусти браузер"], 59, true},                       //11
		{"отключись", idsCommand["отключись"], 100, true},                               //12
		{"включить", idsCommand["отключись"], 82, true},                                 //13
	}
	for idx, c := range cases {
		if i, r, found := assist.FoundCommandByRatio(c.cmd); i != c.i || r != c.r || found != c.found {
			t.Fatalf("case#%v idx=%v ratio=%v found=%v %s", idx, i, r, found, c.cmd)
		}
	}

}

func TestRotateCommand2(t *testing.T) {
	Init()
	type testRotate struct {
		cmd   string
		i     string
		found bool
	}

	cases := []testRotate{
		{"тсет", "", false},                //0
		{"тест", idsCommand["test"], true}, //1
		{"скажи который час", idsCommand["который час"], true},      //2
		{"какое сейчас время", idsCommand["который час"], true},     //3
		{"сколько сейчас времени", idsCommand["который час"], true}, //4
		{"подскажи время", "", false},                               //5
		{"включи", "", false}, //6
		{"включи свет в ванне пожалуйста", idsCommand["включи свет в ванне"], true}, //7
		{"включи свет в ванной", idsCommand["включи свет в ванне"], true},           //8
		{"выключить свет в ванной", idsCommand["выключи свет в ванне"], true},       //9
		{"выключи свет в ванной", idsCommand["выключи свет в ванне"], true},         //10
		{"запусти стим", "", false},                      //11
		{"отключись", idsCommand["отключись"], true},     //12
		{"включить", "", false},                          //13
		{"включи стин", idsCommand["включи стим"], true}, //14
	}
	for idx, c := range cases {
		if i, found := assist.ComparingCommand(c.cmd); i != c.i || found != c.found {
			t.Fatalf("case#%v idx=%v found=%v %s", idx, i, found, c.cmd)
		}
	}

}

func TestIsFindedNameInText(t *testing.T) {
	names := []string{
		"альфа",
		"бета",
	}

	cases := map[string]bool{
		"альфа включи свет":                   true,
		"включи свет":                         false,
		"бета включи свет в ванне пожалуйста": true,
	}

	for text, v := range cases {
		if smarty.IsFindedNameInText(names, text) != v {
			t.Fatalf("case `%s` is FAILED", text)
		}
	}
}

func TestMatchCommand(t *testing.T) {
	ctx := context.TODO()
	recognize := &rcgnz{}
	assist = smarty.New(ctx)
	assist.SetRecognizeCommand(recognize)
	assist.SetRecognizeName(recognize)

	var idsCmd = make(map[string]string)

	idsCmd["0"] = assist.AddCommand([]string{"^поставь будильник на (?P<time>\\d) (?P<range>.*)$"}, func(ctx context.Context, a *smarty.Assiser) {})            //1
	idsCmd["1"] = assist.AddCommand([]string{"^поставь будильник на$"}, func(ctx context.Context, a *smarty.Assiser) {})                                        //2
	idsCmd["2"] = assist.AddCommand([]string{"^какая погода (на|будет)\\s?(?P<date>\\d+)?\\s?(?P<day>\\D+)$"}, func(ctx context.Context, a *smarty.Assiser) {}) //3
	idsCmd["3"] = assist.AddCommand([]string{"какая сейчас погода"}, func(ctx context.Context, a *smarty.Assiser) {})                                           //4
	idsCmd["4"] = assist.AddCommand([]string{"громкость на (?P<volume>\\d+)\\s?(%|процентов)?"}, func(ctx context.Context, a *smarty.Assiser) {})               //5
	idsCmd["5"] = assist.AddCommand([]string{"включи стим"}, func(ctx context.Context, a *smarty.Assiser) {})                                                   //6
	idsCmd["6"] = assist.AddCommand([]string{"отключись", "выключись"}, func(ctx context.Context, a *smarty.Assiser) {})                                        //7
	idsCmd["7"] = assist.AddCommand([]string{"счет", "^счёт$"}, func(ctx context.Context, a *smarty.Assiser) {})                                                //4

	type Variant struct {
		talks   string
		params  map[string]string
		idx     string
		founded bool
	}

	cases := []Variant{
		{
			talks: "поставь будильник на 2 часа",
			params: map[string]string{
				"range": "часа",
				"time":  "2",
			},
			idx:     "0",
			founded: true,
		},
		{
			talks:   "поставь будильник на два часа",
			params:  nil,
			idx:     "",
			founded: false,
		},
		{
			talks: "какая погода на завтра",
			params: map[string]string{
				"day": "завтра",
			},
			idx:     "2",
			founded: true,
		},
		{
			talks: "какая погода будет 14 декабря",
			params: map[string]string{
				"date": "14",
				"day":  "декабря",
			},
			idx:     "2",
			founded: true,
		},
		{
			talks:   "какая сейчас погода",
			params:  map[string]string{},
			idx:     "3",
			founded: true,
		},
		{
			talks:   "какая сейчас погода 123",
			params:  map[string]string{},
			idx:     "3",
			founded: true,
		},
		{
			talks: "громкость на 50 процентов",
			params: map[string]string{
				"volume": "50",
			},
			idx:     "4",
			founded: true,
		},
		{
			talks: "громкость на 50%",
			params: map[string]string{
				"volume": "50",
			},
			idx:     "4",
			founded: true,
		},
		{
			talks: "громкость на 50",
			params: map[string]string{
				"volume": "50",
			},
			idx:     "4",
			founded: true,
		},
		{
			talks: "громкость на пятьдесят процентов",
			params: map[string]string{
				"volume": "50",
			},
			idx:     "",
			founded: false,
		},
		{
			talks:   "какая счёт погода",
			params:  map[string]string{},
			idx:     "",
			founded: false,
		},
	}

	for i, v := range cases {
		idx, params, founded := assist.MatchCommand(v.talks)
		if v.founded != founded || idsCmd[v.idx] != idx {
			t.Fatalf("FAILED case #%v %s idx=%v cmd founded=%v case founded=%v", i, v.talks, v.idx, founded, v.founded)
		}
		fmt.Println("case #", i, v.talks, params)
	}
}

func TestDelete(t *testing.T) {

	ctx := context.TODO()
	recognize := &rcgnz{}
	assist = smarty.New(ctx)
	assist.SetRecognizeCommand(recognize)
	assist.SetRecognizeName(recognize)

	id := ""
	id = assist.AddCommand([]string{"тест"}, func(ctx context.Context, a *smarty.Assiser) {
		defer func() {
			if id != "" {
				c, _ := a.GetCommand(id)
				go func() {
					for {
						if !c.IsActive {
							a.DeleteCommand(id)
							fmt.Println("deleted", id)
							break
						}
						time.Sleep(time.Millisecond * 100)
					}
				}()
			}
		}()
	})

	assist.RunCommand("тест")
	time.Sleep(time.Second * 3)
	fmt.Println(assist.GetCommands())
}

// func TestVoiceCommand(t *testing.T) {

// 	ctx := context.TODO()
// 	recognize := &rcgnz{}
// 	assist = smarty.New(ctx)
// 	assist.SetRecognizeCommand(recognize)
// 	assist.SetRecognizeName(recognize)

// 	assist.AddGenCommand(smarty.ObjectCommand{Commands: []string{"привет"}, Type: "voice", Path: "", Args: []string{"привет"}})

// 	assist.RunCommand("привет")
// 	time.Sleep(time.Second * 2)

// }

// func TestGameMode(t *testing.T) {
// 	godotenv.Load()

// 	log := logger.New("app")
// 	log.LogLevel = logger.INFO

// 	recognizer := voskclient.New()
// 	recognizer.SetLogger(log)
// 	recognizer.Host = os.Getenv("VOSK_HOST")
// 	recognizer.Port = os.Getenv("VOSK_PORT")
// 	ctx := context.TODO()

// 	assist = smarty.New(ctx)
// 	assist.SetConfig(smarty.Config{
// 		Names:           []string{"альфа"},
// 		ListenLongTime:  time.Second / 2,
// 		LenWavBuf:       40,
// 		MaxEmptyMessage: 40,
// 	})
// 	assist.SetLogger(log)
// 	assist.SetRecognizeCommand(recognizer)
// 	assist.SetRecognizeName(recognizer)

// 	assist.SetGameMode(true)
// 	assist.RecognizeEmptyWav = false

// 	assist.Start()

// }

// func TestSetMicrophone(t *testing.T) {
// 	devices, _ := listen.GetMicrophons()
// 	fmt.Println(devices)
// 	ctx, cancel := context.WithCancel(context.Background())
// 	recognize := &rcgnz{}
// 	assist = smarty.New(ctx)
// 	assist.SetRecognizeCommand(recognize)
// 	assist.SetRecognizeName(recognize)
// 	// valid microphone name
// 	assist.MicrophoneName = "Микрофон (High Definition Audio Device)"
// 	go func() {
// 		time.Sleep(time.Second * 2)
// 		cancel()
// 	}()
// 	assist.Start()

// }
