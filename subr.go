// Package subr is a light wrapper over the flag package for sub-command
// handling.
package subr // import "github.com/tr-d/subr"
import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

// Status ...
type Status int

// Statuses
const (
	Safeword Status = iota + 1
	NoArgs
	ParseError
	UnknownSub
)

// Cmd ...
type Cmd struct {
	Name  string         // the name of the sub-command
	Usage string         // a usage string, formatted with 1 argument: os.Args[0]
	Fset  *flag.FlagSet  // like it says on the tin
	Fn    func(*Cmd) int // func which executes command

	Safeword string   // set to "help", for example
	Args     []string // remaining positional arguments
	Svc      Servicer // custom context/services back-door

	Status Status
	Detail string

	flags map[string]*string // string flags
	flagb map[string]*bool   // bool flags
	flagi map[string]*int    // int flags
}

// Servicer ...
type Servicer interface {
	Connect() error
}

// AddFlag ...
func (c *Cmd) AddFlag(flag string, dflt interface{}, usage string) {
	if c.Fset == nil {
		fmt.Fprintln(os.Stderr, "missing FlagSet")
		return
	}
	switch reflect.TypeOf(dflt).String() {
	case "string":
		if len(c.flags) < 1 {
			c.flags = map[string]*string{}
		}
		c.flags[flag] = c.Fset.String(flag, dflt.(string), usage)
	case "bool":
		if len(c.flagb) < 1 {
			c.flagb = map[string]*bool{}
		}
		c.flagb[flag] = c.Fset.Bool(flag, dflt.(bool), usage)
	case "int":
		if len(c.flagi) < 1 {
			c.flagi = map[string]*int{}
		}
		c.flagi[flag] = c.Fset.Int(flag, dflt.(int), usage)
	default:
		fmt.Fprintf(os.Stderr, "unsupported flag type %s\n", reflect.TypeOf(dflt))
		return
	}
}

// Submit ...
func (c *Cmd) Submit() int {
	return c.Fn(c)
}

func (c Cmd) String() string {
	fhelp := []string{"FLAGS"}
	c.Fset.VisitAll(func(f *flag.Flag) {
		fhelp = append(fhelp, fmt.Sprintf("    -%s  %s", f.Name, f.Usage))
	})
	return c.Usage + "\n" + strings.Join(fhelp, "\n")
}

// Parse ...
func Parse(args []string, cmds ...*Cmd) *Cmd {
	if len(args) < 1 {
		return &Cmd{Status: NoArgs}
	}
	for _, cmd := range cmds {
		cmd.Fset.Usage = func() { return }
		cmd.Fset.SetOutput(ioutil.Discard)
		if args[0] == cmd.Name {
			if err := cmd.Fset.Parse(args[1:]); err != nil {
				cmd.Status = ParseError
				cmd.Detail = err.Error()
				return cmd
			}
			cmd.Args = cmd.Fset.Args()
			if len(cmd.Args) > 0 && cmd.Args[0] == cmd.Safeword {
				cmd.Status = Safeword
			}
			return cmd
		}
	}
	return &Cmd{Status: UnknownSub, Name: args[0]}
}
