# go-cluster-ssh: cssh replacement in go

## What are this?

For anyone who needs to administrate clusters of many machines,
[clusterssh](http://sourceforge.net/projects/clusterssh/) has long been a
fallback for when the rest of your automation tools aren't working.

go-cluster-ssh aims to be a simple replacement for cssh with the following improvements:

- Uses a single window to hold multiple terminals.
- Intelligently tiles terminals to fit available window size.
- Scrolls available terminals when they don't all fit in-window.
- Never resizes a terminal smaller than 250x250 characters.
- Uses GTK and the VTE widget to provide modern, anti-aliased terminals.
- Group terminals
- Save and load groups

## Install

The install process is very simple on most distros:

* Download binary from [Release page](https://github.com/Ak-Army/go-cluster-ssh/releases)

Run ```go-cluster-ssh -- HOST [HOST ...]```

## Examples

Basic usage is covered via the builtin help, which you can get by running
```go-cluster-ssh -h```. This section covers some common use cases.

To connect to a list of hosts in a file:
```bash
go-cluster-ssh -f hostlist.txt
```

To use a custom login name, public key, or other SSH client options:
```bash
go-cluster-ssh -args="-l someuser -i ~/.ssh/myotherkey" -- HOST [HOST ...]
go-cluster-ssh -args="-l someuser" -args="-i ~/.ssh/myotherkey" -- HOST [HOST ...]
```

To do something other than ssh, such as edit a bunch of files in parallel:
```bash
go-cluster-ssh -e nano *.txt
```

## Usage Tips

Doing a clustered paste isn't completely obvious. The following methods will
work, after making sure you're clicked into the text entry box at the bottom of
the window:

- middle click or shift-insert to paste the X11 selection buffer.
- control-shift-v to paste the GTK/GNOME clipboard.

## Bugs & To Do

To see current issues, report problems, and see plans for features,
see the [go-cluster-ssh GitHub issues page](https://github.com/Ak-Army/go-cluster-ssh/issues).

## Build dependency
sudo apt install libgtk-3-dev libcairo2-dev libglib2.0-dev libvte-2.91-dev libgirepository1.0-dev


