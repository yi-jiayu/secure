package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const DefaultConfigFileName = "secure.conf"

type Config map[string]string

type ParseError struct {
	LineNumber int
	Line       string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("error parsing config at line %d: %s", e.LineNumber, e.Line)
}

func ParseConfig(r io.Reader) (Config, error) {
	config := make(Config)
	scanner := bufio.NewScanner(r)
	for i := 1; scanner.Scan(); i++ {
		line := scanner.Text()
		// skip empty lines and comments
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) > 1 {
			key, val := fields[0], fields[1]
			config[key] = val
		} else {
			return nil, &ParseError{LineNumber: i, Line: line}
		}
	}

	return config, nil
}

func ParseConfigFile(name string) (Config, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %v", err)
	}
	return ParseConfig(f)
}
