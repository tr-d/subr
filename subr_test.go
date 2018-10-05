package subr

import (
	"flag"
	"testing"
)

// TestSubWithFlags ...
func TestSubWithFlags(t *testing.T) {
	scFset := flag.NewFlagSet("subcmd", flag.ExitOnError)
	c := Cmd{
		Name:  "subcmd",
		Usage: "subcmd help document",
		Fset:  scFset,
		Flags: map[string]*string{
			"string":  scFset.String("string", "", "test string flag"),
			"dstring": scFset.String("dstring", "default", "test string flag, default"),
		},
		Flagb: map[string]*bool{
			"bool":  scFset.Bool("bool", false, "test bool flag"),
			"dbool": scFset.Bool("dbool", true, "test bool flag, default"),
		},
		Safeword: "halp",
	}

	enoArgs := []string{}
	cmd, err := Parse(enoArgs, &c)
	if cmd != nil {
		t.Errorf("expected nil cmd, got %+v\n", cmd)
	}
	switch err := err.(type) {
	case nil:
		t.Error("expected NoArgs error, got nil")
	case *NoArgs:
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
}
