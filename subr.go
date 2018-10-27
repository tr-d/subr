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
	HelpInvoked
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
	switch reflect.TypeOf(dflt).Kind() {
	case reflect.String:
		if len(c.flags) < 1 {
			c.flags = map[string]*string{}
		}
		c.flags[flag] = c.Fset.String(flag, dflt.(string), usage)
	case reflect.Bool:
		if len(c.flagb) < 1 {
			c.flagb = map[string]*bool{}
		}
		c.flagb[flag] = c.Fset.Bool(flag, dflt.(bool), usage)
	case reflect.Int:
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

// S ...
func (c Cmd) S(n string) string {
	if v, ok := c.flags[n]; ok {
		return *v
	}
	return ""
}

// B ...
func (c Cmd) B(n string) bool {
	if v, ok := c.flagb[n]; ok {
		return *v
	}
	return false
}

// I ...
func (c Cmd) I(n string) int {
	if v, ok := c.flagi[n]; ok {
		return *v
	}
	return 0
}

// Pop ...
func (c Cmd) Pop(in interface{}) {
	v := reflect.ValueOf(in).Elem()
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		flag := t.Field(i).Tag.Get("subr")
		if flag == "" {
			continue
		}
		switch f.Type().String() {
		case "string":
			f.SetString(c.S(flag))
		case "bool":
			f.SetBool(c.B(flag))
		case "int":
			f.SetInt(int64(c.I(flag)))
		}
	}
}

func (c Cmd) String() string {
	fhelp := []string{}
	c.Fset.VisitAll(func(f *flag.Flag) {
		fhelp = append(fhelp, fmt.Sprintf("    -%s  %s", f.Name, f.Usage))
	})
	return fmt.Sprintf(c.Usage, strings.Join(fhelp, "\n"))
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
				if err == flag.ErrHelp {
					cmd.Status = HelpInvoked
					return cmd
				}
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
