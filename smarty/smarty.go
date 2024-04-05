package smarty

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	fuzzy "github.com/paul-mannino/go-fuzzywuzzy"
	"github.com/playmixer/corvid/listen"
	"golang.org/x/exp/slices"
)

/**
 TODO
 feature

 - распознование по регулярке
 - передача параметров из команды
 - сеттер разпознования команды

**/

type AssiserEvent int

const (
	AEStartListeningCommand AssiserEvent = 10
	AEStartListeningName    AssiserEvent = 20
	AEApplyCommand          AssiserEvent = 30

	F_MIN_TOKEN    = 90
	F_MAX_DISTANCE = 10
	F_MIN_RATIO    = 75

	LEN_WAV_BUFF = 20

	CMD_MAX_EMPTY_MESSAGE = 10
)

type iLogger interface {
	ERROR(v ...string)
	INFO(v ...string)
	DEBUG(v ...string)
}

type logger struct{}

func (l *logger) ERROR(v ...string) {
	log.Println("ERROR", v)
}

func (l *logger) INFO(v ...string) {
	log.Println("INFO", v)
}

func (l *logger) DEBUG(v ...string) {
	log.Println("DEBUG", v)
}

type CommandFunc func(ctx context.Context, a *Assiser)

type CommandStruct struct {
	Commands []string
	Func     CommandFunc
	IsActive bool
	Context  context.Context
	Cancel   context.CancelFunc
}

type Config struct {
	Names           []string
	ListenLongTime  time.Duration
	LenWavBuf       int
	MaxEmptyMessage int
}

type IReacognize interface {
	Recognize(bufWav []byte) (string, error)
}

type wavBuffer struct {
	buf  *[]byte
	text string
}

type assistentStatus struct {
	LastCommand string
}

type AssitentVoice struct {
	Enable      bool
	ErrorEnable bool
	Func        func(ctx context.Context, text string) error
	Params      map[string]string
}

type Assiser struct {
	ctx              context.Context
	log              iLogger
	Names            []string
	listenLongTime   time.Duration
	lenWavBuf        int
	maxEmptyMessage  int
	commands         map[string]*CommandStruct
	muCmd            sync.Mutex
	eventChan        chan AssiserEvent
	recognizeCommand IReacognize
	recognizeName    IReacognize
	voice            AssitentVoice
	Status           assistentStatus
	wavBuffer        []wavBuffer
	recorder         *listen.Listener
	UserSaid         chan string
	sync.Mutex
}

func New(ctx context.Context) *Assiser {

	a := &Assiser{
		log:             &logger{},
		ctx:             ctx,
		Names:           []string{"альфа", "alpha"},
		listenLongTime:  time.Second * 2,
		lenWavBuf:       LEN_WAV_BUFF,
		maxEmptyMessage: CMD_MAX_EMPTY_MESSAGE,
		commands:        map[string]*CommandStruct{},
		muCmd:           sync.Mutex{},
		eventChan:       make(chan AssiserEvent, 1),
		voice: AssitentVoice{
			Enable:      true,
			ErrorEnable: true,
			Func:        func(ctx context.Context, text string) error { return nil },
			Params:      make(map[string]string),
		},
		wavBuffer: make([]wavBuffer, LEN_WAV_BUFF),
		Status: assistentStatus{
			LastCommand: "",
		},
		UserSaid: make(chan string),
	}

	return a
}

func (a *Assiser) GetContext() context.Context {
	return a.ctx
}

func (a *Assiser) GetRecorder() *listen.Listener {
	return a.recorder
}

func (a *Assiser) addWavToBuf(w *wavBuffer) *[]wavBuffer {
	a.Lock()
	defer a.Unlock()
	a.wavBuffer = append([]wavBuffer{*w}, a.wavBuffer[:a.lenWavBuf-1]...)

	return &a.wavBuffer
}

func (a *Assiser) GetWavFromBuf(count int) []byte {
	a.Lock()
	defer a.Unlock()
	result := *a.wavBuffer[0].buf
	for i := 1; i <= count && i <= a.lenWavBuf && i < len(a.wavBuffer); i++ {
		if a.wavBuffer[i].buf != nil {
			result = listen.ConcatWav(*a.wavBuffer[i].buf, result)
		}
	}
	return result
}

func (a *Assiser) SetRecognizeCommand(recognize IReacognize) {
	a.recognizeCommand = recognize
}

func (a *Assiser) SetRecognizeName(recognize IReacognize) {
	a.recognizeName = recognize
}

func (a *Assiser) GetCommands() [][]string {
	result := [][]string{}
	for i := range a.commands {
		result = append(result, a.commands[i].Commands)
	}

	return result
}

func (a *Assiser) AddCommand(cmd []string, f CommandFunc) string {
	id := RandStringRunes(20) + "_" + time.Now().Format(time.RFC3339Nano)
	a.commands[id] = &CommandStruct{
		Commands: cmd,
		Func:     f,
	}
	return id
}

func (a *Assiser) RunCommand(cmd string) {
	a.Lock()
	defer a.Unlock()
	i, found := a.ComparingCommand(cmd)
	a.log.DEBUG("rotate command", cmd, fmt.Sprint(i), fmt.Sprint(found))
	if found {
		a.Status.LastCommand = cmd
		a.log.DEBUG("Run command", cmd)
		a.PostSignalEvent(AEApplyCommand)
		ctx, cancel := context.WithCancel(a.ctx)
		if _, ok := a.commands[i]; ok {
			a.commands[i].Context = ctx
			a.commands[i].Cancel = cancel
			go func() {
				a.commands[i].IsActive = true
				a.commands[i].Func(ctx, a)
				a.commands[i].IsActive = false
				a.commands[i].Cancel()
			}()
		}
	}

}

func (a *Assiser) RotateCommand(talk string) (index string, percent int) {
	var idx string = ""
	percent = 0
	var founded bool = false
	for i, command := range a.commands {
		for _, c := range command.Commands {
			//проверяем что все слова команды есть в предложении пользователя
			wordsCommand := strings.Fields(c)
			wordsTalk := strings.Fields(talk)
			allWordsInCommand := true
			for _, word := range wordsCommand {
				if ok := slices.Contains(wordsTalk, word); !ok {
					allWordsInCommand = false
					break
				}
			}
			p := fuzzy.TokenSetRatio(c, talk)
			if p > percent && allWordsInCommand {
				idx, percent = i, p
				founded = true
			}
		}
	}
	if !founded {
		return "", 0
	}

	return idx, percent
}

func (a *Assiser) FoundCommandByToken(talk string) (index string, percent int, founded bool) {
	var idx string = ""
	percent = 0
	founded = false
	for i, command := range a.commands {
		for _, c := range command.Commands {
			p := fuzzy.TokenSetRatio(c, talk)
			if p > percent {
				idx, percent = i, p
				founded = true
			}
			// fmt.Printf("%s vs %s, percent=%v, founded=%v\n", talk, c, p, founded)
		}
	}
	return idx, percent, founded
}

func (a *Assiser) FoundCommandByDistance(talk string) (index string, distance int, founded bool) {
	var idx string = ""
	distance = 1000000
	founded = false
	for i, command := range a.commands {
		for _, c := range command.Commands {
			d := fuzzy.EditDistance(c, talk)
			if d < distance {
				idx, distance = i, d
				founded = true
			}
			// fmt.Printf("%s vs %s, distance=%v, founded=%v\n", talk, c, d, founded)
		}
	}
	return idx, distance, founded
}

func (a *Assiser) FoundCommandByRatio(talk string) (index string, ratio int, founded bool) {
	var idx string = ""
	ratio = 0
	founded = false
	for i, command := range a.commands {
		for _, c := range command.Commands {
			r := fuzzy.Ratio(c, talk)
			if r > ratio {
				idx, ratio = i, r
				founded = true
			}
			// fmt.Printf("%s vs %s, ratio=%v, founded=%v\n", talk, c, r, founded)
		}
	}
	return idx, ratio, founded
}

func (a *Assiser) MatchCommand(talk string) (index string, params map[string]string, founded bool) {
	founded = false
	for i, command := range a.commands {
		for _, c := range command.Commands {

			r := regexp.MustCompile(c)
			matches := r.FindStringSubmatch(talk)

			params = make(map[string]string)
			if matches != nil {
				for s, name := range r.SubexpNames() {
					if s > 0 && s <= len(matches) {
						params[name] = matches[s]
					}
				}
			}
			if len(matches) > 0 {
				return i, params, true
			}
		}
	}

	return "", map[string]string{}, false
}

func (a *Assiser) prepareCommand(talk string) string {
	talk = strings.ToLower(talk)
	splitTalk := strings.Split(talk, " ")

	//удаляем имя из команды
	arrCommand := []string{}
	for i := range splitTalk {
		for idxName := range a.Names {
			if splitTalk[i] == a.Names[idxName] {
				continue
			}
		}
		arrCommand = append(arrCommand, splitTalk[i])
	}
	splitTalk = arrCommand

	talk = strings.Join(splitTalk, " ")
	a.log.DEBUG("prepare command: " + talk)
	return talk
}

func (a *Assiser) ComparingCommand(talk string) (index string, found bool) {
	talk = a.prepareCommand(talk)
	found = false
	t, tv, tf := a.FoundCommandByToken(talk)
	d, dv, df := a.FoundCommandByDistance(talk)
	r, rv, rf := a.FoundCommandByRatio(talk)
	if tf == df && df == rf && rf &&
		t == d && d == r &&
		tv >= F_MIN_TOKEN && (dv <= F_MAX_DISTANCE || rv >= F_MIN_RATIO) {
		a.log.DEBUG(fmt.Sprintf("find command id=%v, talk=%s, command=%s", t, talk, a.commands[t].Commands[0]))
		return t, tf
	}
	rIdx, rParams, rFounded := a.MatchCommand(talk)
	if rFounded {
		a.voice.Params = rParams
		a.log.DEBUG(fmt.Sprintf("find command id=%v, talk=%s, command=%s, params=%s", t, talk, a.commands[t].Commands[0], a.voice.Params))
		return rIdx, rFounded
	}

	a.log.DEBUG(fmt.Sprintf("not found command '%s'", talk))
	return "", false
}

func (a *Assiser) SetConfig(cfg Config) {
	a.Names = cfg.Names
	a.listenLongTime = cfg.ListenLongTime
	// a.lenWavBuf = cfg.LenWavBuf
	a.wavBuffer = make([]wavBuffer, cfg.LenWavBuf)
	a.maxEmptyMessage = cfg.MaxEmptyMessage
}

func (a *Assiser) SetLogger(log iLogger) {
	a.log = log
}

func (a *Assiser) InitDefaultCommand() {
	a.AddCommand([]string{"стоп", "stop"}, func(ctx context.Context, a *Assiser) {
		for i := range a.commands {
			if a.commands[i].IsActive {
				a.log.INFO("Стоп", fmt.Sprint(a.commands[i].Commands[0]))
				a.commands[i].Cancel()
			}
		}
	})
}

func (a *Assiser) Start() {
	log := a.log
	if a.recognizeCommand == nil || a.recognizeName == nil {
		log.ERROR("Cannot founded recognize method")
		return
	}

	ctx, cancel := context.WithCancel(a.ctx) //вся программа

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// слушаем имя 1ый поток
	log.INFO("Starting listen. Stram #1 ")
	a.recorder = listen.New(a.listenLongTime)
	a.recorder.SetName("Record")
	a.recorder.SetLogger(log)
	a.recorder.Start(ctx)
	defer a.recorder.Stop()

	a.InitDefaultCommand()

	var notEmptyMessageCounter = 0
	var emptyMessageCounter = 0
	var isListenName = true

waitFor:
	for {
		log.DEBUG("for loop")
		// fmt.Println("emptyMessageCounter", emptyMessageCounter)
		// fmt.Println(a.wavBuffer)
		select {
		case <-sigs:
			cancel()
			break waitFor

		case <-a.ctx.Done():
			cancel()
			break waitFor

		case s := <-a.recorder.WavCh:
			log.DEBUG("smarty read from wav chanel")
			txt, err := a.recognizeName.Recognize(s)
			if err != nil {
				log.ERROR(err.Error())
			}
			a.addWavToBuf(&wavBuffer{buf: &s, text: txt})
			if txt != "" {
				notEmptyMessageCounter += 1
				emptyMessageCounter = 0
			}
			if txt == "" {
				if notEmptyMessageCounter > 0 && !isListenName {
					wavB := a.GetWavFromBuf(notEmptyMessageCounter)
					translateText, err := a.recognizeCommand.Recognize(wavB)
					if err != nil {
						log.ERROR(err.Error())
					}
					log.INFO("Вы сказали: " + translateText)
					a.userSaid(translateText)
					a.RunCommand(translateText)
				}
				if emptyMessageCounter > a.maxEmptyMessage && !isListenName {
					isListenName = true
					a.PostSignalEvent(AEStartListeningName)
				}
				notEmptyMessageCounter = 0
				emptyMessageCounter += 1
			}
			if isListenName && len(a.wavBuffer) > 1 && notEmptyMessageCounter > 1 {
				wavB := a.GetWavFromBuf(2)
				textWithName, err := a.recognizeName.Recognize(wavB)
				if err != nil {
					log.ERROR(err.Error())
				}
				if IsFindedNameInText(a.Names, textWithName) {
					isListenName = false
					a.PostSignalEvent(AEStartListeningCommand)
				}

			}
		}
	}
}

func (a *Assiser) Print(t ...any) {
	fmt.Println(a.Names[0]+": ", fmt.Sprint(t...))
}

func (a *Assiser) PostSignalEvent(s AssiserEvent) {
	select {
	case a.eventChan <- s:
	default:
	}

}

func (a *Assiser) GetSignalEvent() <-chan AssiserEvent {
	return a.eventChan
}

func (a *Assiser) SetTTS(f func(ctx context.Context, text string) error) {
	a.voice.Func = f
}

func (a *Assiser) Voice(ctx context.Context, text string) error {
	if !a.voice.Enable {
		return nil
	}

	return a.voice.Func(ctx, text)
}

func (a *Assiser) VoiceError(ctx context.Context, text string) error {
	if !a.voice.ErrorEnable {
		return nil
	}

	return a.Voice(ctx, text)
}

type TypeCommand string

const (
	tcExec TypeCommand = "exec"
	tcTool TypeCommand = "tool"
)

type ObjectCommand struct {
	Type     TypeCommand `json:"type"`
	Path     string      `json:"path"`
	Args     []string    `json:"args"`
	Commands []string    `json:"commands"`
}

/**
* создаем команду для запуска внешнего процесса
 */
func (a *Assiser) newCommandExec(pathFile string, args ...string) CommandFunc {

	return func(ctx context.Context, a *Assiser) {
		go func() {
			err := exec.Command(pathFile, args...).Run()
			if err != nil {
				a.log.ERROR(pathFile + " " + strings.Join(args, " ") + " error: " + err.Error())
			}
		}()
	}
}

func (a *Assiser) LoadCommands(filepath string) error {
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		return err
	}
	cByte, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	var data []ObjectCommand
	err = json.Unmarshal(cByte, &data)
	if err != nil {
		return err
	}

	for i := range data {
		a.AddGenCommand(data[i])
	}
	return nil
}

func (a *Assiser) AddGenCommand(data ObjectCommand) {
	a.log.DEBUG(fmt.Sprintf("add command %s %s %s %s", strings.Join(data.Commands, "|"), data.Type, data.Path, strings.Join(data.Args, "|")))
	var f CommandFunc
	if data.Type == tcExec {
		f = a.newCommandExec(data.Path, data.Args...)
	}
	if data.Type == tcTool {
		f = a.newCommandTool(data.Path, data.Args)
	}
	if f != nil {
		a.AddCommand(data.Commands, f)
	}
}

func (a *Assiser) DeleteAllCommand() {
	a.commands = make(map[string]*CommandStruct)
}

func (a *Assiser) DeleteCommand(id string) {
	delete(a.commands, id)
}

func (a *Assiser) GetCommand(id string) (*CommandStruct, error) {
	if _, ok := a.commands[id]; ok {
		return a.commands[id], nil
	}

	return nil, fmt.Errorf("not found command by id: %s", id)
}

func (a *Assiser) userSaid(txt string) {
	select {
	case a.UserSaid <- txt:
	default:
	}
}

func IsFindedNameInText(names []string, text string) bool {
	for _, name := range names {
		if fuzzy.TokenSetRatio(name, text) == 100 {
			return true
		}
	}
	return false
}
