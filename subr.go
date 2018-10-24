// Package subr is a light wrapper over the flag package for sub-command
// handling.
package subr // import "github.com/tr-d/subr"
import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
)

// Cmd ...
type Cmd struct {
	Name  string        // the name of the sub-command
	Usage string        // a usage string, formatted with 1 argument: os.Args[0]
	Fset  *flag.FlagSet // like it says on the tin

	Safeword string   // set to "help", for example
	Args     []string // remaining positional arguments
	Svc      Servicer // custom context/services back-door

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

func (c Cmd) String() string {
	return fmt.Sprintf(c.Usage, os.Args[0])
}

// Parse ...
func Parse(args []string, cmds ...*Cmd) (*Cmd, error) {
	if len(args) < 1 {
		return nil, &NoArgs{}
	}
	for _, cmd := range cmds {
		if cmd.Name == "" {
			return nil, &ParseFailed{errors.New("missing command name")}
		}
		cmd.Fset.Usage = func() { return }
		cmd.Fset.SetOutput(ioutil.Discard)
		if args[0] == cmd.Name {
			if err := cmd.Fset.Parse(args[1:]); err != nil {
				return nil, &ParseFailed{err}
			}
			cmd.Args = cmd.Fset.Args()
			if len(cmd.Args) > 0 && cmd.Args[0] == cmd.Safeword {
				return nil, &Safeword{}
			}
			return cmd, nil
		}
	}
	return nil, &UnknownSub{args[0]}
}

// ERRORS

// NoArgs is an error.
type NoArgs struct{}

func (e *NoArgs) Error() string {
	return "no args to parse"
}

// ParseFailed is an error.
type ParseFailed struct {
	err error
}

func (e *ParseFailed) Error() string {
	return e.err.Error()
}

// Safeword is an error.
type Safeword struct{}

func (e *Safeword) Error() string {
	return "safeword invoked"
}

// UnknownSub is an error.
type UnknownSub struct{ val string }

func (e *UnknownSub) Error() string {
	return e.val
}
