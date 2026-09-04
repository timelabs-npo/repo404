package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/timelabs-cpo/gccmp/internal/gccmp"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "snapshot":
		err = runSnapshot(os.Args[2:])
	case "compare":
		err = runCompare(os.Args[2:])
	case "verify":
		err = runVerify(os.Args[2:])
	case "version":
		fmt.Println(version)
		return
	case "help", "-h", "--help":
		usage()
		return
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "gccmp:", err)
		os.Exit(2)
	}
}

func runSnapshot(args []string) error {
	fs := flag.NewFlagSet("snapshot", flag.ContinueOnError)
	label := fs.String("label", "root", "stable logical label; absolute source path is never serialized")
	out := fs.String("out", "-", "output JSON path or - for stdout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: gccmp snapshot [--label name] [--out file] ROOT")
	}
	env, err := gccmp.SnapshotDirectory(fs.Arg(0), *label, *out)
	if err != nil {
		return err
	}
	return gccmp.WriteCanonical(*out, env)
}

func runCompare(args []string) error {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	out := fs.String("out", "-", "output JSON path or - for stdout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 2 {
		return fmt.Errorf("usage: gccmp compare [--out file] LEFT.json RIGHT.json")
	}
	left, err := gccmp.ReadSnapshot(fs.Arg(0))
	if err != nil {
		return err
	}
	right, err := gccmp.ReadSnapshot(fs.Arg(1))
	if err != nil {
		return err
	}
	env, err := gccmp.CompareSnapshots(left, right)
	if err != nil {
		return err
	}
	return gccmp.WriteCanonical(*out, env)
}

func runVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: gccmp verify SNAPSHOT.json")
	}
	if err := gccmp.VerifySnapshot(fs.Arg(0)); err != nil {
		return err
	}
	fmt.Println("VERIFIED")
	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, `GCCmp alpha — deterministic, read-only directory observation and comparison

Commands:
  gccmp snapshot --label NAME --out snapshot.json ROOT
  gccmp compare --out comparison.json LEFT.json RIGHT.json
  gccmp verify snapshot.json
  gccmp version

Alpha v0 never copies, deletes, renames, hydrates, uploads, or mutates source trees.`)
}
