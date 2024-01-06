package smarty

import (
	"context"
	"fmt"
	"testing"
)

var (
	assist *Assiser
)

type rcgnz struct{}

func (r *rcgnz) Recognize(bufWav []byte) (string, error) {
	return "тест", nil
}

func Init() {
	ctx := context.TODO()
	recognize := &rcgnz{}
	assist = New(ctx)
	assist.SetRecognizeCommand(recognize)
	assist.SetRecognizeName(recognize)

	assist.AddCommand([]string{"тест"}, func(ctx context.Context, a *Assiser) {})                                          //0
	assist.AddCommand([]string{"который час", "какое время", "сколько времени"}, func(ctx context.Context, a *Assiser) {}) //1
	assist.AddCommand([]string{"включи свет в ванне", "включи в ванной свет"}, func(ctx context.Context, a *Assiser) {})   //2
	assist.AddCommand([]string{"выключи свет в ванне", "выключи в ванной свет"}, func(ctx context.Context, a *Assiser) {}) //3
	assist.AddCommand([]string{"запусти браузер"}, func(ctx context.Context, a *Assiser) {})                               //4
	assist.AddCommand([]string{"включи стим"}, func(ctx context.Context, a *Assiser) {})                                   //5
	assist.AddCommand([]string{"отключись", "выключись"}, func(ctx context.Context, a *Assiser) {})                        //6
}

func TestRotateCommand(t *testing.T) {
	Init()
	type testRotate struct {
		cmd string
		i   int
		p   int
	}

	cases := []testRotate{
		{"тсет", 0, 0},                             //0
		{"тест", 0, 100},                           //1
		{"скажи который час", 1, 100},              //2
		{"какое сейчас время", 1, 100},             //3
		{"сколько сейчас времени", 1, 100},         //4
		{"подскажи время", 0, 0},                   //5
		{"включи", 0, 0},                           //6
		{"включи свет в ванне пожалуйста", 2, 100}, //7
		{"включи свет в ванной", 2, 100},           //8
		{"выключить свет в ванной", 0, 0},          //9
		{"выключи свет в ванной", 3, 100},          //10
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
		i     int
		p     int
		found bool
	}

	cases := []testRotate{
		{"тсет", 0, 75, true},                            //0
		{"тест", 0, 100, true},                           //1
		{"скажи который час", 1, 100, true},              //2
		{"какое сейчас время", 1, 100, true},             //3
		{"сколько сейчас времени", 1, 100, true},         //4
		{"подскажи время", 1, 64, true},                  //5
		{"включи", 2, 100, true},                         //6
		{"включи свет в ванне пожалуйста", 2, 100, true}, //7
		{"включи свет в ванной", 2, 100, true},           //8
		{"выключить свет в ванной", 3, 95, true},         //9
		{"выключи свет в ванной", 3, 100, true},          //10
		{"запусти стим", 4, 74, true},                    //11
		{"отключись", 6, 100, true},                      //12
		{"включить", 6, 82, true},                        //13
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
		i     int
		d     int
		found bool
	}

	cases := []testRotate{
		{"тсет", 0, 2, true},                            //0
		{"тест", 0, 0, true},                            //1
		{"скажи который час", 1, 6, true},               //2
		{"какое сейчас время", 1, 7, true},              //3
		{"сколько сейчас времени", 1, 7, true},          //4
		{"подскажи время", 1, 7, true},                  //5
		{"включи", 6, 3, true},                          //6
		{"включи свет в ванне пожалуйста", 2, 11, true}, //7
		{"включи свет в ванной", 2, 2, true},            //8
		{"выключить свет в ванной", 3, 4, true},         //9
		{"выключи свет в ванной", 3, 2, true},           //10
		{"запусти стим", 5, 6, true},                    //11
		{"отключись", 6, 0, true},                       //12
		{"включить", 6, 2, true},                        //13
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
		i     int
		r     int
		found bool
	}

	cases := []testRotate{
		{"тсет", 0, 75, true},                           //0
		{"тест", 0, 100, true},                          //1
		{"скажи который час", 1, 79, true},              //2
		{"какое сейчас время", 1, 76, true},             //3
		{"сколько сейчас времени", 1, 81, true},         //4
		{"подскажи время", 1, 64, true},                 //5
		{"включи", 6, 80, true},                         //6
		{"включи свет в ванне пожалуйста", 2, 78, true}, //7
		{"включи свет в ванной", 2, 92, true},           //8
		{"выключить свет в ванной", 3, 88, true},        //9
		{"выключи свет в ванной", 3, 93, true},          //10
		{"запусти стим", 4, 59, true},                   //11
		{"отключись", 6, 100, true},                     //12
		{"включить", 6, 82, true},                       //13
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
		i     int
		found bool
	}

	cases := []testRotate{
		{"тсет", 0, false},                          //0
		{"тест", 0, true},                           //1
		{"скажи который час", 1, true},              //2
		{"какое сейчас время", 1, true},             //3
		{"сколько сейчас времени", 1, true},         //4
		{"подскажи время", 0, false},                //5
		{"включи", 0, false},                        //6
		{"включи свет в ванне пожалуйста", 2, true}, //7
		{"включи свет в ванной", 2, true},           //8
		{"выключить свет в ванной", 3, true},        //9
		{"выключи свет в ванной", 3, true},          //10
		{"запусти стим", 0, false},                  //11
		{"отключись", 6, true},                      //12
		{"включить", 0, false},                      //13
		{"включи стин", 5, true},                    //14
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
		if IsFindedNameInText(names, text) != v {
			t.Fatalf("case `%s` is FAILED", text)
		}
	}
}

func TestMatchCommand(t *testing.T) {
	ctx := context.TODO()
	recognize := &rcgnz{}
	assist = New(ctx)
	assist.SetRecognizeCommand(recognize)
	assist.SetRecognizeName(recognize)

	assist.AddCommand([]string{"^поставь будильник на (?P<time>\\d) (?P<range>.*)$"}, func(ctx context.Context, a *Assiser) {})            //1
	assist.AddCommand([]string{"^поставь будильник на$"}, func(ctx context.Context, a *Assiser) {})                                        //2
	assist.AddCommand([]string{"^какая погода (на|будет)\\s?(?P<date>\\d+)?\\s?(?P<day>\\D+)$"}, func(ctx context.Context, a *Assiser) {}) //3
	assist.AddCommand([]string{"какая сейчас погода"}, func(ctx context.Context, a *Assiser) {})                                           //4
	assist.AddCommand([]string{"громкость на (?P<volume>\\d+)\\s?(%|процентов)?"}, func(ctx context.Context, a *Assiser) {})               //5
	assist.AddCommand([]string{"включи стим"}, func(ctx context.Context, a *Assiser) {})                                                   //6
	assist.AddCommand([]string{"отключись", "выключись"}, func(ctx context.Context, a *Assiser) {})                                        //7
	assist.AddCommand([]string{"счет", "^счёт$"}, func(ctx context.Context, a *Assiser) {})                                                //4

	type Variant struct {
		talks   string
		params  map[string]string
		idx     int
		founded bool
	}

	cases := []Variant{
		{
			talks: "поставь будильник на 2 часа",
			params: map[string]string{
				"range": "часа",
				"time":  "2",
			},
			idx:     0,
			founded: true,
		},
		{
			talks:   "поставь будильник на два часа",
			params:  nil,
			idx:     0,
			founded: false,
		},
		{
			talks: "какая погода на завтра",
			params: map[string]string{
				"day": "завтра",
			},
			idx:     2,
			founded: true,
		},
		{
			talks: "какая погода будет 14 декабря",
			params: map[string]string{
				"date": "14",
				"day":  "декабря",
			},
			idx:     2,
			founded: true,
		},
		{
			talks:   "какая сейчас погода",
			params:  map[string]string{},
			idx:     3,
			founded: true,
		},
		{
			talks:   "какая сейчас погода 123",
			params:  map[string]string{},
			idx:     3,
			founded: true,
		},
		{
			talks: "громкость на 50 процентов",
			params: map[string]string{
				"volume": "50",
			},
			idx:     4,
			founded: true,
		},
		{
			talks: "громкость на 50%",
			params: map[string]string{
				"volume": "50",
			},
			idx:     4,
			founded: true,
		},
		{
			talks: "громкость на 50",
			params: map[string]string{
				"volume": "50",
			},
			idx:     4,
			founded: true,
		},
		{
			talks: "громкость на пятьдесят процентов",
			params: map[string]string{
				"volume": "50",
			},
			idx:     0,
			founded: false,
		},
		{
			talks:   "какая счёт погода",
			params:  map[string]string{},
			idx:     0,
			founded: false,
		},
	}

	for i, v := range cases {
		idx, params, founded := assist.MatchCommand(v.talks)
		if v.founded != founded || v.idx != idx {
			t.Fatalf("FAILED case #%v %s idx=%v founded=%v", i, v.talks, idx, founded)
		}
		fmt.Println("case #", i, v.talks, params)
	}
}
