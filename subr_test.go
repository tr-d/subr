package subr

import (
	"flag"
	"fmt"
	"strings"
	"testing"
)

// TestSubWithFlags ...
func TestSubWithFlags(t *testing.T) {
	helpString := "%s subcmd help document"
	scFset := flag.NewFlagSet("subcmd", flag.ContinueOnError)
	c := Cmd{
		Name:  "subcmd",
		Usage: helpString,
		Fset:  scFset,
		Flags: map[string]*string{
			"string":  scFset.String("string", "", "test string flag"),
			"dstring": scFset.String("dstring", "default", "test string flag, default"),
		},
		Flagb: map[string]*bool{
			"bool":  scFset.Bool("bool", false, "test bool flag"),
			"dbool": scFset.Bool("dbool", true, "test bool flag, default"),
		},
		Flagi: map[string]*int{
			"int": scFset.Int("int", 3, "test int flag"),
		},
		Safeword: "halp",
	}

	eName := []string{"subcmd"}
	c.Name = ""
	cmd, err := Parse(eName, &c)
	if cmd != nil {
		t.Errorf("expected nil cmd, got %+v\n", cmd)
	}
	switch err := err.(type) {
	case nil:
		t.Error("expected ParseFailed error, got nil")
	case *ParseFailed:
		if fmt.Sprint(err) != "missing command name" {
			t.Error("bad error message for missing name")
		}
	default:
		t.Errorf("expected ParseFailed error, got %+v\n", err)
	}
	c.Name = "subcmd"

	eParse := []string{"subcmd", "-b", "foo", "bar", "-a"}
	cmd, err = Parse(eParse, &c)
	if cmd != nil {
		t.Errorf("expected nil cmd, got %+v\n", cmd)
	}
	switch err := err.(type) {
	case nil:
		t.Error("expected ParseFailed error, got nil")
	case *ParseFailed:
	default:
		t.Errorf("expected ParseFailed error, got %+v\n", err)
	}

	enoArgs := []string{}
	cmd, err = Parse(enoArgs, &c)
	if cmd != nil {
		t.Errorf("expected nil cmd, got %+v\n", cmd)
	}
	switch err := err.(type) {
	case nil:
		t.Error("expected NoArgs error, got nil")
	case *NoArgs:
		if fmt.Sprint(err) != "no args to parse" {
			t.Error("bad error message for no args")
		}
	default:
		t.Errorf("expected NoArgs error, got %+v\n", err)
	}

	eSafe := []string{"subcmd", "halp", "beep", "beep"}
	cmd, err = Parse(eSafe, &c)
	if cmd != nil {
		t.Errorf("expected nil cmd, got %+v\n", cmd)
	}
	switch err := err.(type) {
	case nil:
		t.Error("expected Safeword error, got nil")
	case *Safeword:
		if fmt.Sprint(err) != "safeword invoked" {
			t.Error("bad error message for safeword")
		}
	default:
		t.Errorf("expected Safeword error, got %+v\n", err)
	}

	eUnkn := []string{"wat", "waaat", "no", "nooo"}
	cmd, err = Parse(eUnkn, &c)
	if cmd != nil {
		t.Errorf("expected nil cmd, got %+v\n", cmd)
	}
	switch err := err.(type) {
	case nil:
		t.Error("expected UnknownSub error, got nil")
	case *UnknownSub:
		if fmt.Sprint(err) != "wat" {
			t.Error("bad error message for unknown sub")
		}
	default:
		t.Errorf("expected UnknownSub error, got %+v\n", err)
	}

	args := []string{"subcmd", "-string", "value", "-bool", "pos1", "pos2"}
	cmd, err = Parse(args, &c)
	if err != nil {
		t.Errorf("got parse error: %v\n", err)
	}
	if len(cmd.Args) != 2 {
		t.Errorf("expected 2 positional args, got: %v\n", cmd.Args)
	}
	if len(cmd.Flags) != 2 {
		t.Errorf("expected two string flags, got %d\n", len(cmd.Flags))
	}
	if len(cmd.Flagb) != 2 {
		t.Errorf("expected two bool flags, got %d\n", len(cmd.Flagb))
	}
	if *cmd.Flags["string"] != "value" {
		t.Errorf("expected -string value, got \"-string %v\"\n", *cmd.Flags["string"])
	}
	if *cmd.Flags["dstring"] != "default" {
		t.Errorf("expected -dstring default, got \"-dstring %v\"\n", *cmd.Flags["dstring"])
	}
	if !*cmd.Flagb["bool"] {
		t.Error("expected -bool true, got false")
	}
	if !*cmd.Flagb["dbool"] {
		t.Error("expected -dbool true, got false")
	}
	if strings.HasSuffix(fmt.Sprint(cmd), helpString) {
		t.Error("expected print(cmd) to be correct, is not")
	}
}
