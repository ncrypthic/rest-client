package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	wmenu "github.com/dixonwille/wmenu/v5"
	"github.com/ncrypthic/rest-client/parser/text"
)

const (
	ActionExecute int = 1 << iota
	ActionView
	ActionEdit
	ActionDelete
	ActionList
)

type Endpoint struct {
	Source  string
	Request *http.Request
}

type VariableContext struct {
	Variable *text.Variable
}

type EndpointContext struct {
	Endpoint *Endpoint
	Variable *text.Variable
	Segment  string
	Action   int
}

func (cmd *EndpointContext) SetRequest(opt wmenu.Opt) error {
	switch t := opt.Value.(type) {
	case *Endpoint:
		cmd.Endpoint = t
	case *text.Variable:
		cmd.Variable = t
	default:
		return errors.New("Unhandled expectation")
	}
	return nil
}

func (cmd *EndpointContext) SetAction(opt wmenu.Opt) error {
	action := opt.Value.(int)
	switch action {
	case ActionExecute:
		return cmd.Execute(cmd.Endpoint)
	case ActionView:
		return cmd.Display(cmd.Endpoint, cmd.Variable)
	default:
	}
	return nil
}

func (cmd *EndpointContext) Execute(endpoint *Endpoint) error {
	res, err := http.DefaultClient.Do(cmd.Endpoint.Request)
	if err != nil {
		log.Print(err.Error())
	} else {
		text, _ := ioutil.ReadAll(res.Body)
		for k, v := range res.Header {
			fmt.Printf("%s: %s\n", k, strings.Join(v, ";"))
		}
		fmt.Println()
		fmt.Printf("%s\n", string(text))
	}
	fmt.Fscanf(os.Stdin, "Press any key")
	return err
}

func (cmd *EndpointContext) Edit(opt wmenu.Opt) error {
	return nil
}

func (cmd *EndpointContext) Display(endpoint *Endpoint, variable *text.Variable) error {
	if cmd.Endpoint != nil {
		cmd.DisplayEndpoint(cmd.Endpoint)
	} else if cmd.Variable != nil {
		cmd.DisplayVariable(variable)
	}
	return nil
}

func (cmd *EndpointContext) DisplayVariable(v *text.Variable) error {
	credentials := ""
	port := ""
	if v.URL.User != nil {
		credentials = v.URL.String()
	}
	if v.URL.Port() != "" {
		port = fmt.Sprintf(":%s", v.URL.Port())
	}
	fmt.Printf("%s://%s%s%s\n", v.URL.Scheme, credentials, v.URL.Hostname(), port)
	fmt.Println()
	for k, v := range v.Header {
		fmt.Printf("%s: %s\n", k, strings.Join(v, ";"))
	}
	fmt.Fscanf(os.Stdin, "Press any key")
	return nil
}

func (cmd *EndpointContext) DisplayEndpoint(endpoint *Endpoint) error {
	fmt.Println(endpoint.Source)
	fmt.Fscanf(os.Stdin, "Press any key")
	return nil
}

type MenuHandler func(wmenu.Opt) error

type Builder func(MenuHandler) *wmenu.Menu

type ChainMenu func(*wmenu.Menu) MenuHandler

func After(handler MenuHandler) ChainMenu {
	return func(next *wmenu.Menu) MenuHandler {
		return func(opt wmenu.Opt) error {
			err := handler(opt)
			if err == nil {
				return next.Run()
			}
			return err
		}
	}
}

func MainMenu(requests []*http.Request, segments []string, variable *text.Variable, cmd *EndpointContext) Builder {
	endpointMenu := wmenu.NewMenu("Choose endpoint")
	endpointMenu.ChangeReaderWriter(os.Stdin, os.Stdout, os.Stderr)
	endpointMenu.ClearOnMenuRun()
	return func(handler MenuHandler) *wmenu.Menu {
		endpointMenu.Option("Variable", variable, false, func(opt wmenu.Opt) error {
			variableMenu := wmenu.NewMenu("Variable menu")
			variableMenu.ClearOnMenuRun()
			variableMenu.Option("View", nil, true, func(_ wmenu.Opt) error {
				cmd.DisplayVariable(variable)
				return nil
			})
			return variableMenu.Run()
		})
		for idx, request := range requests {
			endpoint := &Endpoint{
				Request: request,
				Source:  segments[idx],
			}
			endpointMenu.Option(fmt.Sprintf("%s %s", request.Method, request.URL.String()), endpoint, false, handler)
		}
		return endpointMenu
	}
}

func VariableMenu(variable *text.Variable) Builder {
	variableMenu := wmenu.NewMenu("Variable menu")
	return func(handler MenuHandler) *wmenu.Menu {
		variableMenu.Option("View", ActionView, true, handler)
		return variableMenu
	}
}

func ActionMenu() Builder {
	actionMenu := wmenu.NewMenu("Endpoint menu")
	actionMenu.ClearOnMenuRun()
	return func(handler MenuHandler) *wmenu.Menu {
		actionMenu.Option("View", ActionView, true, handler)
		actionMenu.Option("Execute", ActionExecute, false, handler)
		return actionMenu
	}
}

func watchFile(path string) (<-chan bool, error) {
	ch := make(chan bool)
	initialStat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	go func(filePath string, initialStat os.FileInfo) {
		for {
			stat, err := os.Stat(filePath)
			if err != nil {
				ch <- true
			} else if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
				log.Println("File changed")
				initialStat = stat
				select {
				case ch <- true:
				default:
				}
			}

			time.Sleep(500 * time.Millisecond)
		}
	}(path, initialStat)
	return ch, nil
}

func application(path string, ch <-chan bool) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	requests, segments, variable, err := text.Parse(data)
	if err != nil {
		return err
	}
	go func() {
		for {
			<-ch
			log.Println("Reload endpoints")
			data, err := ioutil.ReadFile(path)
			if err == nil {
				requests, segments, variable, _ = text.Parse(data)
			}
		}
	}()
	for {
		ctx := &EndpointContext{}
		mainMenu := MainMenu(requests, segments, variable, ctx)
		actionMenu := ActionMenu()(ctx.SetAction)
		mainMenu(
			After(ctx.SetRequest)(actionMenu),
		).Run()
	}
}

func main() {
	chFile, err := watchFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	application(os.Args[1], chFile)
}
