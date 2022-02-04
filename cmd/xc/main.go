package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"

	"github.com/joe-davidson1802/xc/parser"
	"github.com/posener/complete"
)

var (
	version = ""
)

func completion(fileName string) bool {
	cmp := complete.New("xc", complete.Command{
		GlobalFlags: complete.Flags{
			"-version": complete.PredictNothing,
			"-h":       complete.PredictNothing,
			"-short":   complete.PredictNothing,
			"-help":    complete.PredictNothing,
			"-f":       complete.PredictFiles("*.md"),
			"-file":    complete.PredictFiles("*.md"),
		},
	})
	b, err := os.Open(fileName)
	if err == nil {
		p, err := parser.NewParser(b)
		if err != nil {
			return false
		}
		t, err := p.Parse()
		if err != nil {
			return false
		}
		s := make(map[string]complete.Command)
		for _, ta := range t {
			s[ta.Name] = complete.Command{}
		}
		cmp.Command.Sub = s
	}
	cmp.CLI.InstallName = "complete"
	cmp.CLI.UninstallName = "uncomplete"
	cmp.AddFlags(nil)

	flag.Parse()

	return cmp.Complete()
}

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	flag.Usage = func() {
		fmt.Println("xc - list tasks")
		fmt.Println("xc [task...] - run tasks")
		flag.PrintDefaults()
	}

	var (
		versionFlag bool
		helpFlag    bool
		fileName    string
		short       bool
		md          bool
	)

	flag.BoolVar(&versionFlag, "version", false, "show xc version")
	flag.BoolVar(&helpFlag, "help", false, "shows xc usage")
	flag.BoolVar(&helpFlag, "h", false, "shows xc usage")
	flag.StringVar(&fileName, "file", "README.md", "specify markdown file that contains tasks")
	flag.StringVar(&fileName, "f", "README.md", "specify markdown file that contains tasks")
	flag.BoolVar(&short, "short", false, "list task names in a short format")
	flag.BoolVar(&md, "md", false, "print the markdown for a task rather than running it")

	if completion(fileName) {
		return
	}

	b, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	p, err := parser.NewParser(b)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	t, err := p.Parse()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if versionFlag {
		fmt.Printf("xc version: %s\n", getVersion())
		return
	}

	if helpFlag {
		flag.Usage()
		return
	}
	tav := getArgs()
	if len(tav) == 0 && !short {
		fmt.Println("tasks:")
		maxLen := 0
		for _, n := range t {
			if len(n.Name) > maxLen {
				maxLen = len(n.Name)
			}
		}
		for _, n := range t {
			padLen := maxLen - len(n.Name)
			pad := strings.Repeat(" ", padLen)
			var desc []string
			if n.ParsingError != "" {
				desc = append(desc, fmt.Sprintf("Parsing Error: %s", n.ParsingError))
			}
			for _, d := range n.Description {
				desc = append(desc, fmt.Sprintf("%s", d))
			}
			if len(n.DependsOn) > 0 {
				desc = append(desc, fmt.Sprintf("Requires:  %s", strings.Join(n.DependsOn, ", ")))
			}
			if len(desc) == 0 {
				desc = append(desc, n.Commands...)
			}
			fmt.Printf("    %s%s  %s\n", n.Name, pad, desc[0])
			for _, d := range desc[1:] {
				fmt.Printf("    %s  %s\n", strings.Repeat(" ", maxLen), d)
			}
		}
		return
	}
	if len(tav) == 0 && short {
		for _, n := range t {
			fmt.Println(n.Name)
		}
		return
	}
	if md {
		if len(tav) != 1 {
			fmt.Printf("md requires 1 task, got: %d\n", len(tav))
			os.Exit(1)
		}
		ta, ok := t.Get(tav[0])
		if !ok {
			fmt.Printf("%s is not a task\n", tav[0])
		}
		ta.Display(os.Stdout)
		return

	}
	for _, tav := range tav {
		err = t.ValidateDependencies(tav, []string{})
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		err = t.Run(context.Background(), tav)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

}
func getArgs() []string {
	var (
		args = flag.Args()
	)
	return args
}

func getVersion() string {
	if version != "" {
		return version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" {
		return "unknown"
	}

	version = info.Main.Version
	if info.Main.Sum != "" {
		version += fmt.Sprintf(" (%s)", info.Main.Sum)
	}

	return version
}
