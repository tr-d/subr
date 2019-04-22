// Package subr is a light wrapper over the flag package for sub-command
// handling.
package subr // import "github.com/tr-d/subr"
import (
	"bufio"
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
	Name     string // the name of the sub-command
	Usage    string // a usage string, formatted with 1 argument: os.Args[0]
	Safeword string // default from New(): help

	Fset *flag.FlagSet  // like it says on the tin
	Fn   func(*Cmd) int // func which executes command
	Svc  Servicer       // custom context/services back-door

	Args []string // remaining positional arguments

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

// New ...
func New(name string, usage string, fn func(*Cmd) int) *Cmd {
	return &Cmd{
		Name:     name,
		Usage:    usage,
		Fn:       fn,
		Fset:     flag.NewFlagSet(name, flag.ContinueOnError),
		Safeword: "help",
	}
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
	if c.Fn == nil {
		return -1
	}
	return c.Fn(c)
}

// Bind ...
func (c Cmd) Bind(in interface{}) {
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

// HasStdin returns true if stdin is a char device.
func (c Cmd) HasStdin() bool {
	f, err := os.Stdin.Stat()
	if err != nil {
		return false // TODO this should maybe be a panic?
	}
	return f.Mode()&os.ModeCharDevice == 0
}

// HasPipe returns true if stdout is a pipe.
func (c Cmd) HasPipe() bool {
	f, err := os.Stdout.Stat()
	if err != nil {
		return false // TODO this should maybe be a panic?
	}
	return f.Mode()&os.ModeNamedPipe != 0
}

// ReadStdin will return a slice of bytes from stdin.
func (c Cmd) ReadStdin() []byte {
	b := []byte{}
	if !c.HasStdin() {
		return b
	}
	r := bufio.NewReader(os.Stdin)
	b, _ = ioutil.ReadAll(r)
	return b
}

// StdinArgs returns a stdin as a slice of strings split on newlines.
func (c Cmd) StdinArgs() []string {
	a := []string{}
	if !c.HasStdin() {
		return a
	}
	b := c.ReadStdin()
	return strings.Split(strings.TrimSpace(string(b)), "\n")
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

func (c Cmd) String() string {
	// TODO better formatting, preserve flag order (from .AddFlag())
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
