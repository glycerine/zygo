package zygo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// read new-line delimited text from a file into an array (slurpf "path-to-file")
func SlurpfileFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	var fn string
	switch fna := args[0].(type) {
	case SexpStr:
		fn = string(fna)
	default:
		return SexpNull, fmt.Errorf("slurp requires a string path to read. we got type %T / value = %v", args[0], args[0])
	}

	if !FileExists(string(fn)) {
		return SexpNull, fmt.Errorf("file '%s' does not exists", fn)
	}
	f, err := os.Open(fn)
	if err != nil {
		return SexpNull, err
	}
	defer f.Close()

	a := make([]Sexp, 0)

	bufIn := bufio.NewReader(f)
	lineNum := int64(1)
	for {
		lastline, err := bufIn.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return SexpNull, err
		}
		n := len(lastline)
		if err == io.EOF && n == 0 {
			break
		}
		if n > 0 {
			if lastline[n-1] == '\n' {
				a = append(a, SexpStr(string(lastline[:n-1])))
			} else {
				a = append(a, SexpStr(string(lastline)))
			}
			lineNum += 1
		}

		if err == io.EOF {
			break
		}
	}

	VPrintf("read %d lines\n", lineNum)
	return SexpArray(a), nil
}

// (writef path); (write path) is the macro version.
// (owritef path): write an array of strings out to the named file,
// overwriting it in the process. (owrite) is the macro version.
func WriteToFileFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}
	var fn string
	switch fna := args[1].(type) {
	case SexpStr:
		fn = string(fna)
	default:
		return SexpNull, fmt.Errorf("owrite requires a string (SexpStr) path to write to as the first argument. we got type %T / value = %v", args[1], args[1])
	}

	if name == "write" || name == "writef" {
		// don't overwrite existing file
		if FileExists(fn) {
			return SexpNull, fmt.Errorf("refusing to write to existing file '%s'",
				fn)
		}
	}
	// owrite / owritef overwrite indiscriminately.

	f, err := os.Create(fn)
	if err != nil {
		return SexpNull, err
	}
	defer f.Close()

	var slice []Sexp
	switch sl := args[0].(type) {
	case SexpArray:
		slice = []Sexp(sl)
		for i := range slice {
			s := slice[i].SexpString()
			if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
				s = s[1 : len(s)-1]
			}
			_, err = fmt.Fprintf(f, "%s\n", s)
			if err != nil {
				return SexpNull, err
			}
		}
	default:
		s := sl.SexpString()
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			s = s[1 : len(s)-1]
		}
		_, err = fmt.Fprintf(f, "%s\n", s)
		if err != nil {
			return SexpNull, err
		}
	}

	return SexpNull, nil
}

// SplitStringFunction splits a string based on an arbitrary delimiter
func SplitStringFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 2 {
		return SexpNull, WrongNargs
	}

	// make sure the two args are strings
	s1, ok := args[0].(SexpStr)
	if !ok {
		return SexpNull, fmt.Errorf("split requires a string to split, got %T", args[0])
	}
	s2, ok := args[1].(SexpStr)
	if !ok {
		return SexpNull, fmt.Errorf("split requires a string as a delimiter, got %T", args[1])
	}

	toSplit := string(s1)
	splitter := string(s2)
	s := strings.Split(toSplit, splitter)

	split := make([]Sexp, len(s))
	for i := range split {
		split[i] = SexpStr(s[i])
	}

	return SexpArray(split), nil
}

// (nsplit "a\nb") -> ["a" "b"]
func SplitStringOnNewlinesFunction(env *Glisp, name string, args []Sexp) (Sexp, error) {
	if len(args) != 1 {
		return SexpNull, WrongNargs
	}
	args = append(args, SexpStr("\n"))

	return SplitStringFunction(env, name, args)
}
