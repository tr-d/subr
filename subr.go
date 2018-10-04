// Package subr is a light wrapper over the flag package for sub-command
// handling.
package subr // import "github.com/tr-d/subr"
import (
	"errors"
	"flag"
)

// Cmd ...
type Cmd struct {
	Name  string        // the name of the sub-command
	Usage string        // a usage string, formatted with 1 argument: os.Args[0]
	Exec  func(Cmd) int // the function which executes the sub-command
	Fset  *flag.FlagSet // like it says on the tin

	Flags map[string]*string // string flags
	Flagb map[string]*bool   // bool flags
	Flagi map[string]*int    // int flags

	Safeword string   // set to "help", for example
	Args     []string // remaining positional arguments
	Svc      Servicer // custom context/services back-door
}

// Servicer ...
type Servicer interface {
	Connect() error
}

// Parse ...
func Parse(args []string, cmds ...*Cmd) (*Cmd, error) {
	if len(args) < 1 {
		return nil, &NoArgs{}
	}
	for _, cmd := range cmds {
		if cmd.Name == nil {
			return nil, &ParseFailed{errors.New("missing command name")}
		}
		if args[0] == *cmd.Name {
			if err := cmd.fset.Parse(args[1:]); err != nil {
				return nil, &ParseFailed{err}
			}
			cmd.Args = fset.Args()
			if len(cmd.Args) > 0 && cmd.Args[0] == cmd.Safeword {
				return nil, &Safeword{}
			}
			return nil, cmd
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
