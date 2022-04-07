package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {

	f, err := os.Open("instruction.go")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read instruction.go: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	codeDefinition := regexp.MustCompile(`^\s+.+\s+=\sInstructionCode\(".+"\)`)
	funcDefinition := regexp.MustCompile(`^\s+.+:\sfunc\(cpu\s\*CPU,\sraw\suint64\)\sException\s\{$`)

	codes := map[string]struct{}{}
	fns := map[string]struct{}{}

	for scanner.Scan() {
		line := scanner.Text()
		if codeDefinition.Match([]byte(line)) {
			code := strings.TrimSpace(strings.Split(line, "=")[0])
			if code == "_INVALID" {
				continue
			}
			if strings.HasPrefix(code, "//") {
				continue
			}
			codes[code] = struct{}{}
			continue
		}

		if funcDefinition.Match([]byte(line)) {
			fn := strings.TrimSpace(strings.Split(line, ":")[0])
			fns[fn] = struct{}{}
			continue
		}
	}

	if err = scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "scan: %v", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "unimplemented codes:\n")
	for code := range codes {
		_, ok := fns[code]
		if !ok {
			fmt.Fprintf(os.Stdout, "%s\n", code)
		}
	}
}
