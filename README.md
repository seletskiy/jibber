Jibber is simple backend which can be used with Jira WebHook to send
notifications about changes in issues directly to XMPP room.

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

Example configuration to send jabber notifications
---------------------------------------------------

### Configure jibber

`/etc/jibber/jibber.conf`:

```
mod_rest

-u
  http://your-ejabberd-server:5280/rest/

-t
  your-target-room@bla.bla.bla
```

Launch jibber to listen (default port is 65432).

### Configure Jira

Go to Administration/System/WebHooks and add WebHook, like this:
![WebHook example](https://cloud.githubusercontent.com/assets/674812/4677167/eb99fdc6-55e3-11e4-9e87-3ecab3f32651.png)

After that, you should see notifications in your specified room when issues in
specified project are changed.
