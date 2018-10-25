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
	return c.Usage
}

// Parse ...
func Parse(args []string, cmds ...*Cmd) (*Cmd, int, error) {
	if len(args) < 1 {
		return nil, 2, errors.New("no args")
	}
	for _, cmd := range cmds {
		if cmd.Name == "" {
			return nil, 3, errors.New("missing command.Name")
		}
		cmd.Fset.Usage = func() { return }
		cmd.Fset.SetOutput(ioutil.Discard)
		if args[0] == cmd.Name {
			if err := cmd.Fset.Parse(args[1:]); err != nil {
				return nil, 4, err
			}
			cmd.Args = cmd.Fset.Args()
			if len(cmd.Args) > 0 && cmd.Args[0] == cmd.Safeword {
				return cmd, 1, errors.New("safeword invoked")
			}
			return cmd, 0, nil
		}
	}
	return nil, 5, fmt.Errorf("unknown subcmd: %s", args[0])
}
