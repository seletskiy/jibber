package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
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

	"github.com/mattn/go-xmpp"
	"github.com/seletskiy/tplutil"
)

var configPath = "/etc/jibber/jibber.conf"

var reIndent = regexp.MustCompile(`(?m)^`)

type webHookHandler struct {
	tplDir  string
	mainTpl string
	output  io.Writer
	debug   bool
}

type stdoutOutput struct{}

type xmppCommon struct {
	to   string
	join bool
	nick string
	opts xmpp.Options
	talk *xmpp.Client
	tpl  map[string]*template.Template
}

func (s stdoutOutput) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

type ejabberModRest struct {
	url  string
	from string
	to   string
}

func (output ejabberModRest) Write(p []byte) (n int, err error) {
	xmlNode := struct {
		XMLName xml.Name `xml:"message"`
		from    string   `xml:"from,attr"`
		to      string   `xml:"to,attr"`
		body    []byte   `xml:"body"`
	}{
		from: output.from,
		to:   output.to,
		body: p,
	}

	msg, err := xml.Marshal(&xmlNode)
	if err != nil {
		log.Println("error while marshalling msg to XML:", err)
		return 0, err
	}

	http.Post(output.url, "application/xml", bytes.NewBuffer(msg))

	return len(p), nil
}

func (output *xmppCommon) Connect() error {
	talk, err := output.opts.NewClient()
	if err != nil {
		return err
	}

	go func() {
		for {
			if _, err := talk.Recv(); err != nil {
				// just ignore everything
				return
			}
		}
	}()

	output.talk = talk

	if output.join {
		talk.JoinMUCNoHistory(output.to, output.nick)
		log.Printf("xmpp: joined <%s> as <%s>\n", output.to, output.nick)
	}

	return nil
}

func (output *xmppCommon) Write(p []byte) (n int, err error) {
	msgType := "chat"
	if output.join {
		msgType = "groupchat"
	}

	return output.talk.Send(xmpp.Chat{
		Remote: output.to,
		Type:   msgType,
		Text:   string(p),
	})
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
		template.New(h.mainTpl).Funcs(tplutil.Last).Funcs(tplFuncs),
		filepath.Join(h.tplDir, "*.tpl"),
	)
	if err != nil {
		log.Println("error while parsing template:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if h.debug {
		prettyJson, _ := json.MarshalIndent(result, "", "  ")
		fmt.Fprint(os.Stderr, string(prettyJson)+"\n")
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

	if output, ok := h.output.(*xmppCommon); ok && !output.join {
		output.to, err = tplutil.ExecuteToString(output.tpl["normal"], result)
		if err != nil {
			log.Println("error while executing tpl:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	_, err = h.output.Write([]byte(msg))
	if err != nil {
		log.Printf("error writing to output:", err)
		if _, ok := h.output.(*xmppCommon); ok {
			if err := h.output.(*xmppCommon).Connect(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Fatal(err)
			} else {
				log.Printf("reconnected")
				_, err = h.output.Write([]byte(msg))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Fatal(err)
				}
			}
		}
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
			url:  args["--url"].(string),
			from: args["--from"].(string),
			to:   args["--to"].(string),
		}
	case args["xmpp"]:
		noTLS := args["--start-tls"].(bool) || args["--no-tls"].(bool)
		log.Printf("%#v", args)

		statusMsg := ""
		if args["--status-msg"] != nil {
			statusMsg = args["--status-msg"].(string)
		}

		xmppOutput := &xmppCommon{
			to:   args["--to"].(string),
			join: args["--join"].(bool),
			nick: args["--nick"].(string),
			opts: xmpp.Options{
				Host:          args["--host"].(string),
				User:          args["--user"].(string),
				Password:      args["--pass"].(string),
				NoTLS:         noTLS,
				Debug:         args["--debug"].(bool),
				StartTLS:      args["--start-tls"].(bool),
				Session:       true,
				Status:        args["--status"].(string),
				StatusMessage: statusMsg,
			},
		}

		xmpp.DefaultConfig = tls.Config{
			ServerName:         strings.Split(xmppOutput.opts.Host, ":")[0],
			InsecureSkipVerify: args["--no-verify-tls-host"].(bool),
		}

		err := xmppOutput.Connect()
		if err != nil {
			log.Fatal(err)
		}

		if args["--presence"] != nil {
			_, err := xmppOutput.Write([]byte(args["--presence"].(string)))
			if err != nil {
				log.Fatal(err)
			}
		}

		if !xmppOutput.join {
			xmppOutput.tpl = make(map[string]*template.Template)
			xmppOutput.tpl["normal"], err = template.New("normal").Parse(xmppOutput.to)
			if err != nil {
				log.Println("error while parsing template:", err)
				return
			}
		}

		output = xmppOutput
	}

	http.Handle("/", webHookHandler{
		tplDir:  args["--tpl-dir"].(string),
		mainTpl: args["--tpl"].(string),
		output:  output,
		debug:   args["--debug"].(bool),
	})

	log.Println("listening on", args["-l"])
	log.Println(http.ListenAndServe(args["-l"].(string), nil))
}

func parseArgs() map[string]interface{} {
	usage := `Jira to Jabber Notification Bridge.

You should specify 'backend' as first option. Supported are:
	* stdout - just print formed message to stdout, for debug.
	* mod_rest - use ejabberd mod_rest module (does not support MUC, but doesn't require auth).
	* xmpp - full blown XMPP client (support MUC).

Usage:
  jibber [options] stdout
  jibber [options] xmpp --host HOSTNAME --user USERNAME --pass PASSWORD --to SEND-TO [--join]
  jibber [options] mod_rest --url MOD-REST-URL --to SEND-TO [--from SEND-FROM]

Options:
  -h --help             Show this help message.
  --url MOD-REST-URL    {mod_rest} Use ejabberd mod_rest URL [default: http://localhost:5280/rest].
  --from SEND-FROM      {mod_rest} Jabber ID send from (can be any) [default: jira-notifier].
  --to SEND-TO          {mod_rest,xmpp} Jabber ID send message to (most useful with rooms).
                        It can also be a template.
  --host HOSTNAME       {xmpp} Jabber server hostname.
  --user USERNAME       {xmpp} Username to log in.
  --pass PASSWORD       {xmpp} Password for that username.
  --no-tls              {xmpp} Do not use TLS [default: false].
  --no-verify-tls-host  {xmpp} Do not verify certificate hostname [default: false].
  --start-tls           {xmpp} Use STARTTLS if server support it [default: false].
  --nick NICK           {xmpp} Use nick in MUC [default: Jira].
  --join                {xmpp} Join to room (specified as --to) [default: false].
  --debug               {xmpp} Display debug information (XML) to stdout [default: false].
  --status STATUS       {xmpp} Status to set [default: online].
  --status-msg MSG      {xmpp} Status message to set.
  --presence MSG        {xmpp} Send presence message when connect is successfull.
                        It will also print raw JSON request from Jira.
  -l LISTEN-ADDR        HTTP addr:port to listen to [default: :65432].
  --tpl-dir DIR         Template dir to form messages [default: /etc/jibber/tpl].
  --tpl TPL-NAME        Main template to start from [default: main.tpl].`

	rawArgs := make([]string, 0)
	conf, err := ioutil.ReadFile(configPath)
	if err == nil {
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

	args, _ := docopt.Parse(usage, rawArgs, true, "jibber v1.1", false)

	return args
}
