// 2023 FMNX team.
// Use of this code is governed by GNU General Public License.
// Official web page: https://fmnx.su/core/pack
// Contact email: help@fmnx.su

package main

import (
	"fmt"
	"os"
	"strings"

	"fmnx.su/core/pack/pack"
	"fmnx.su/core/pack/pacman"
	"fmnx.su/core/pack/tmpl"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	Help    bool `long:"help" short:"h"`
	Version bool `long:"version" short:"v"`

	// Root options.
	Query  bool `short:"Q" long:"query"`
	Remove bool `short:"R" long:"remove"`
	Sync   bool `short:"S" long:"sync"`
	Push   bool `short:"P" long:"push"`
	Build  bool `short:"B" long:"build"`

	// Sync options.
	Quick     bool   `short:"q" long:"quick"`
	Refresh   []bool `short:"y" long:"refresh"`
	Upgrade   []bool `short:"u" long:"upgrade"`
	Info      []bool `short:"i" long:"info"`
	List      []bool `short:"l" long:"list"`
	Notimeout bool   `short:"j" long:"notimeout"`
	Force     bool   `short:"f" long:"force"`
	Keepcfg   bool   `short:"k" long:"keepcfg"`

	// Push options.
	Dir      string `short:"d" long:"dir" default:"/var/cache/pacman/pkg"`
	Insecure bool   `short:"w" long:"insecure"`
	Endpoint string `long:"endpoint" default:"/api/packages/arch"`
	Distro   string `long:"distro" default:"archlinux"`

	// Remove options.
	Confirm     bool `short:"o" long:"confirm"`
	Norecursive bool `long:"norecursive"`
	Nocfgs      bool `long:"nocfgs"`
	Cascade     bool `long:"cascade"`

	// Query options.
	Explicit bool   `long:"explicit"`
	Unreq    bool   `long:"unreq"`
	File     string `long:"file"`
	Foreign  bool   `long:"foreign"`
	Deps     bool   `long:"deps"`
	Native   bool   `long:"native"`
	Groups   bool   `long:"groups"`
	Check    []bool `long:"check"`

	// Build options.
	Syncbuild bool `short:"s" long:"syncbuild"`
	Rmdeps    bool `short:"r" long:"rmdeps"`
	Garbage   bool `short:"g" long:"garbage"`
	Template  bool `short:"t" long:"template"`
	ExportKey bool `short:"e" long:"exp-key"`
}

func main() {
	_, err := flags.NewParser(&opts, flags.None).Parse()
	CheckErr(err)

	switch {
	case opts.Sync && opts.Help:
		fmt.Println(tmpl.SyncHelp)
		return

	case opts.Sync:
		CheckErr(pack.Sync(args(), pack.SyncParameters{
			Quick:     opts.Quick,
			Refresh:   opts.Refresh,
			Upgrade:   opts.Upgrade,
			Info:      opts.Info,
			List:      opts.List,
			Notimeout: opts.Notimeout,
			Force:     opts.Force,
			Keepcfg:   opts.Keepcfg,
			Insecure:  opts.Insecure,
			Stdout:    os.Stdout,
			Stderr:    os.Stderr,
			Stdin:     os.Stdin,
		}))
		return

	case opts.Push && opts.Help:
		fmt.Println(tmpl.PushHelp)
		return

	case opts.Push:
		CheckErr(pack.Push(args(), pack.PushParameters{
			Directory: opts.Dir,
			Insecure:  opts.Insecure,
			Endpoint:  opts.Endpoint,
			Stdout:    os.Stdout,
			Stderr:    os.Stderr,
			Stdin:     os.Stdin,
		}))
		return

	case opts.Remove && opts.Help:
		fmt.Println(tmpl.RemoveHelp)
		return

	case opts.Remove:
		CheckErr(pacman.RemoveList(args(), pacman.RemoveParameters{
			Sudo:        true,
			NoConfirm:   !opts.Confirm,
			Recursive:   !opts.Norecursive,
			WithConfigs: !opts.Nocfgs,
			Stdout:      os.Stdout,
			Stderr:      nil,
			Stdin:       os.Stdin,
		}))
		return

	case opts.Query && opts.Help:
		fmt.Println(tmpl.QueryHelp)
		return

	case opts.Query:
		CheckErr(pacman.Query(args(), pacman.QueryParameters{
			Explicit:   opts.Explicit,
			Deps:       opts.Deps,
			Native:     opts.Native,
			Foreign:    opts.Foreign,
			Unrequired: opts.Unreq,
			Groups:     opts.Groups,
			Info:       opts.Info,
			Check:      opts.Check,
			List:       opts.List,
			File:       opts.File,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
			Stdin:      os.Stdin,
		}))
		return

	case opts.Build && opts.Help:
		fmt.Println(tmpl.BuildHelp)
		return

	case opts.Build:
		CheckErr(pack.Build(pack.BuildParameters{
			Dir:       opts.Dir,
			Quick:     opts.Quick,
			Syncbuild: opts.Syncbuild,
			Rmdeps:    opts.Rmdeps,
			Garbage:   opts.Garbage,
			Template:  opts.Template,
			ExportKey: opts.ExportKey,
			Stdout:    os.Stdout,
			Stderr:    os.Stderr,
			Stdin:     os.Stdin,
		}))
		return

	case opts.Version:
		fmt.Println(tmpl.Version)
		return

	case opts.Help:
		fmt.Println(tmpl.Help)
		return

	default:
		fmt.Println("Please, specify at least one root flag (pack -h)")
		os.Exit(1)
		return
	}
}

// Herlper function to exit on unexpected errors.
func CheckErr(err error) {
	if err != nil {
		fmt.Println(tmpl.Err + err.Error())
		os.Exit(1)
	}
}

// This gets list of all arguements and removes command, string args and bool
// args from list. New string arguements should be added to stringargs variable
// for command to work properly.
// TODO: later rewrite with reflect to avoid unexpected behaviour.
func args() []string {
	var stringargs = []string{
		"-n", "--name", "-p", "--port", "--cert", "--key", "--ring", "--file",
		"--protocol", "--endopint", "-d", "--dir", "--arch", "--distro",
	}
	var filtered []string
	for i, v := range os.Args {
		if i == 0 || i == 1 {
			continue
		}
		if strings.HasPrefix(v, "-") {
			continue
		}
		var next bool
		for _, args := range stringargs {
			if os.Args[i-1] == args {
				next = true
			}
		}
		if next {
			continue
		}
		filtered = append(filtered, v)
	}
	return filtered
}
