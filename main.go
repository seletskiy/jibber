package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/docopt/docopt-go"
	"github.com/seletskiy/tplutil"
)

var configPath = "/etc/jibber/jibber.conf"

var reIndent = regexp.MustCompile(`(?m)^`)

type webHookHandler struct {
	TplDir  string
	MainTpl string
	Output  io.Writer
}

type stdoutOutput struct{}

func (s stdoutOutput) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

type ejabberModRest struct {
	Url  string
	From string
	To   string
}

type ejabberdMsg struct {
	XMLName xml.Name `xml:"message"`
	From    string   `xml:"from,attr"`
	To      string   `xml:"to,attr"`
	Body    []byte   `xml:"body"`
}

func (output ejabberModRest) Write(p []byte) (n int, err error) {
	xmlNode := ejabberdMsg{From: output.From, To: output.To, Body: p}

	msg, err := xml.Marshal(&xmlNode)
	if err != nil {
		log.Println("error while marshalling msg to XML:", err)
		return 0, err
	}

	http.Post(output.Url, "application/xml", bytes.NewBuffer(msg))

	return len(p), nil
}

var tplFuncs = template.FuncMap{
	"indent": func(amount int, val string) string {
		return reIndent.ReplaceAllString(val, strings.Repeat(" ", amount))
	},
	"hasTag": func(tagName string, val string) bool {
		var reHasTag = regexp.MustCompile(`(\W|^)` + tagName + `(\W|$)`)
		return reHasTag.MatchString(val)
	},
}

func (h webHookHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	result := map[string]interface{}{}
	err := json.NewDecoder(req.Body).Decode(&result)
	if err != nil {
		log.Println("error while decoding JSON from WebHook:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tpl, err := tplutil.ParseGlob(
		template.New(h.MainTpl).Funcs(tplutil.Last).Funcs(tplFuncs),
		filepath.Join(h.TplDir, "*.tpl"),
	)
	if err != nil {
		log.Println("error while parsing template:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	msg, err := tplutil.ExecuteToString(tpl, result)
	if err != nil {
		log.Println("error while executing tpl:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if msg == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = h.Output.Write([]byte(msg))
	if err != nil {
		log.Println("error writing to output:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	args := parseArgs()

	var output io.Writer
	switch {
	case args["stdout"]:
		output = stdoutOutput{}
	case args["mod_rest"]:
		output = ejabberModRest{
			Url:  args["-u"].(string),
			From: args["-f"].(string),
			To:   args["-t"].(string),
		}
	}

	http.Handle("/", webHookHandler{
		TplDir:  args["--tpl-dir"].(string),
		MainTpl: args["--tpl"].(string),
		Output:  output,
	})

	log.Println("listening on", args["-l"])
	log.Println(http.ListenAndServe(args["-l"].(string), nil))
}

func parseArgs() map[string]interface{} {
	usage := `Jira to Jabber Notification Bridge.

Only ejabberd mod_rest currently supported because of it's simplest way
to send jabber message. You should consider using it.
	
Usage:
  jibber [options] stdout
  jibber [options] mod_rest -u MOD-REST-URL -t SEND-TO [-f SEND-FROM]

Options:
  -h --help         Show this help message.
  -u MOD-REST-URL   Use ejabberd mod_rest URL [default: http://localhost:5280/rest].
  -t SEND-TO        Jabber ID send message to (most useful with rooms).
  -f SEND-FROM      Jabber ID send from (can be any) [default: jira-notifier].
  -l LISTEN-ADDR    HTTP addr:port to listen to [default: :65432].
  --tpl-dir DIR     Template dir to form messages [default: ./tpl].
  --tpl TPL-NAME    Main template to start from [default: main.tpl].`

	rawArgs := make([]string, 0)
	conf, err := ioutil.ReadFile(configPath)
	if err == nil {
		log.Println("args are read from", configPath)
		confLines := strings.Split(string(conf), "\n")
		for _, line := range confLines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			rawArgs = append(rawArgs, line)
		}
	}

	rawArgs = append(rawArgs, os.Args[1:]...)

	args, _ := docopt.Parse(usage, rawArgs, true, "jibber v1.0", false)

	return args
}
