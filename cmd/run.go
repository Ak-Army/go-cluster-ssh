package cmd

import (
	"bufio"
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/Ak-Army/cli"
	"github.com/sgreben/flagvar"

	"github.com/Ak-Army/go-cluster-ssh/internal"
)

type run struct {
	SSH     string          `flag:"e, specify the SSH executable to use"`
	Args    flagvar.Strings `flag:"args, specify the SSH agruments"`
	Verbose string          `flag:"v, verbose"`
	File    flagvar.Strings `flag:"f, a file containing a list of hosts to connect to"`
	hosts   []*internal.HostGroup
}

func init() {
	cli.RootCommand().AddCommand("run", &run{
		SSH: "/usr/bin/ssh",
	})
}

func (c *run) Help() string {
	return `Usage: go-cluster-ssh [OPTIONS] [HOST ...]`
}

func (c *run) Synopsis() string {
	return "Connect to multiple hosts in parallel."
}

func (c *run) Parse(args []string) error {
	if len(args) > 0 {
		c.hosts = append(c.hosts, &internal.HostGroup{
			Name:  "Default",
			Hosts: args,
		})
	}
	return nil
}

func (c *run) Run(ctx context.Context) error {
	// load hosts from file, if available
	c.handleFiles()
	if len(c.hosts) == 0 {
		return errors.New("no hosts found")
	}
	internal.New(c.hosts, c.SSH, c.Args.Values)

	return nil
}

func (c *run) handleFiles() {
	for _, f := range c.File.Values {
		// hostnames are assumed to be whitespace separated
		file, err := os.Open(f)
		if err != nil {
			log.Fatal(err)
		}
		host := &internal.HostGroup{
			Name: filepath.Base(f),
		}
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			host.Hosts = append(host.Hosts, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		file.Close()
		c.hosts = append(c.hosts, host)
	}
}
