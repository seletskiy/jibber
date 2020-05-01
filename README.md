Jibber is simple tool which can be used with any WebHook (for example, with Jira WebHooks) to transform
hook payload to jabber messages. Very useful with Jira. Just in case you do not want HipChat.

Installation
============

Arch Linux
----------

You can obtain package by checkout `pkgbuild` branch and running `makepkg`:

```
git clone git@github.com:seletskiy/jibber.git
cd jibber
git checkout pkgbuild
makepkg -f
```

Go Get
------

You can always `go get` this package:

```
go get github.com/seletskiy/jibber
```

Configuration
=============

Configuration can be either specified by command line, or via
`/etc/jibber/jibber.conf` file, which has following syntax:

```
<option>
  [<value>]
```

Jibber expects to find some template files in the `/etc/jibber/tpl/`, so
either install `jibber` as package or provide another `--tpl-dir` command
line argument. Like this:

```
jibber --tpl-dir=./tpl/
```

Templates
=========

jibber uses extensive use of golang templates, so no code recompilcation and
no daemon restart is needed to change style of reported messages.

Template syntax is slightly enhanced, consider read about it:
https://godoc.org/github.com/seletskiy/tplutil

jibber will pass JSON fields hierarchy as is to the template. You
can use `--debug` flag to see what's coming from Jira in pretty-pring format
on the stderr.

See default message formats and how they are formatted in the `tpl/` dir.

Example configuration to send jabber notifications
---------------------------------------------------

### Configure jibber
#### Using usual xmpp

`/etc/jibber/jibber.conf`

```
xmpp

--user
    bot@your.host.name

--pass
    bot-password

--to
    target-room@conference.your.host.name

--host
    your.host.name:5222

--join
```

If you see strange errors about TLS and stuff try to use one of following
options:

* `--no-tls` --- disable TLS completely, not recommended, server may not
support that mode;
* `--start-tls` --- use TLS but only when server say about it;
* `--no-verify-tls-host` --- skip hostname validation in certificate check;
* `--debug` --- you will see complete dump of answers from server in XML;

#### Using mod_rest

`/etc/jibber/jibber.conf`

```
mod_rest

--url
    http://your-ejabberd-server:5280/rest/

--to
    target-room@conference.your.host.name
```

#### For debugging

`/etc/jibber/jibber.conf`

```
stdout

--debug
```

Launch jibber to listen (default port is 65432).

#### Test using curl
You can use curl to test your message templates
```
curl --header "Content-Type: application/json" \
  --request POST \
  --data "{\"webhookEvent\": \"example\", \"msg\": \"foobar\"}" \
  http://localhost:65432
```

### Configure Jira

Go to Administration/System/WebHooks and add WebHook, like this:

![WebHook example](https://cloud.githubusercontent.com/assets/674812/4693573/ea51b96c-57a1-11e4-9f4f-6ea45f4a749e.png)

After that, you should see notifications in your specified room when issues in
specified project are changed.

Contributors
============

* Egor Kovetskiy (@kovetskiy)
* Anton Schubert (@iSchluff)
