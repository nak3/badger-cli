/*
 * Copyright 2017 Dgraph Labs, Inc. and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/* badger_info

Usage: badger_cli --dir x [--value-dir y]

This command prints information about the badger key-value store.  It reads MANIFEST and prints its
info. It also prints info about missing/extra files, and general information about the value log
files (which are not referenced by the manifest).  Use this tool to report any issues about Badger
to the Dgraph team.
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/dgraph-io/badger"
)

type cliStateEnum int

const (
	cliStart cliStateEnum = iota
	cliStop
)

func main() {
	dirFlag := flag.String("dir", "", "The Badger database's index directory")
	valueDirFlag := flag.String("value-dir", "",
		"The Badger database's value log directory, if different from the index directory")
	flag.Usage = func() {
		fmt.Printf("Usage:\n\n")
		fmt.Printf("  %s [OPTIONS] [SYNTAX]\n\n", os.Args[0])
		printCliHelp()
		fmt.Printf("  (nil)                     no syntax starts interactive shell")
		fmt.Printf("\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(0)
	}
	flag.Parse()
	err := runCli(*dirFlag, *valueDirFlag, flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printCliHelp() {
	fmt.Print(`Syntax:
  get <KEY>                 get value by key
  set <KEY> <VALUE> <META>  set item by key and value
  delete <KEY>              delete item by key
  dump                      dump item list
`)
}

func printInteractiveHelp() {
	printCliHelp()
	fmt.Print(`Type:
  \q to exit                (Ctrl+C/Ctrl+D also supported)
  \? or "help"              print this help.
`)
}

func runSingleStatement(kv *badger.KV, stmts []string) cliStateEnum {
	switch stmts[0] {
	case "get":
		if len(stmts) != 2 {
			printCliHelp()
			return cliStart
		}
		key := []byte(stmts[1])
		var item badger.KVItem
		if err := kv.Get(key, &item); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v", err)
			return cliStart
		}
		if len(item.Value()) != 0 {
			fmt.Printf("%s\n", item.Value())
		}
	case "set":
		if len(stmts) != 3 {
			printCliHelp()
			return cliStart
		}
		key := []byte(stmts[1])
		value := []byte(stmts[2])
		//userMeta := byte(stmts[3])
		if err := kv.Set(key, value, 0x00); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v", err)
		}
	case "delete":
		if len(stmts) != 2 {
			printCliHelp()
			return cliStart
		}
		key := []byte(stmts[1])
		if err := kv.Delete(key); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v", err)
		}
	case "dump":
		if len(stmts) != 1 {
			printCliHelp()
			return cliStart
		}
		opt := badger.DefaultIteratorOptions
		itr := kv.NewIterator(opt)
		defer itr.Close()
		for itr.Rewind(); itr.Valid(); itr.Next() {
			item := itr.Item()
			key := item.Key()
			val := item.Value() // This could block while value is fetched from value log.
			fmt.Printf("%s %s\n", key, val)
		}
	default:
		printCliHelp()
	}
	return cliStart
}

func runStatements(kv *badger.KV, stmts []string) cliStateEnum {
	if len(stmts) == 0 {
		return cliStart
	}
	switch stmts[0] {
	case `\q`, `exit`:
		return cliStop
	case `\?`, `\help`:
		printInteractiveHelp()
		return cliStart
	default:
		return runSingleStatement(kv, stmts)
	}
	return cliStart
}

func runInteractive(kv *badger.KV, config *readline.Config) error {
	cfg, err := readline.NewEx(config)
	if err != nil {
		return err
	}
	state := cliStart
	for {
		if state == cliStop {
			break
		}
		var err error
		line, err := cfg.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		switch state {
		case cliStart:
			stmts := strings.Fields(strings.TrimSpace(line))
			if len(stmts) > 0 {
				state = runStatements(kv, stmts)
			}
		case cliStop:
			break
		}
	}
	return nil
}

func runCli(dir, valueDir string, args []string) error {
	opt := badger.DefaultOptions
	if dir == "" {
		return fmt.Errorf("--dir not supplied")
	}
	if valueDir == "" {
		valueDir = dir
	}

	opt.Dir = dir
	opt.ValueDir = valueDir

	kv, err := badger.NewKV(&opt)
	if err != nil {
		return err
	}
	defer kv.Close()
	if len(args) > 0 {
		runSingleStatement(kv, args)
		return nil
	}

	config := &readline.Config{
		Prompt:      fmt.Sprintf("%s > ", opt.Dir),
		HistoryFile: "/tmp/readline.tmp",
		//		AutoComplete:    completer,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	}
	return runInteractive(kv, config)
}
