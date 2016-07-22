package main

import(
	"regexp"
	"errors"
	"strings"
	"encoding/json"
)

type(
	Format struct {
		Value   string
		Options string
		parser  Parser
	}

	Parser interface {
		setup(options string) error
		load(logLine []byte)
		parse() error
		clear()
		addField(name string, value interface{})
		getMessage() map[string]interface{}
	}

	Line struct {
		line    []byte
		message map[string]interface{}
	}

	NullLine struct {
		Line
	}

	JsonLine struct {
		Line
	}

	RegexpLine struct {
		Line
		regexp  *regexp.Regexp
	}
)

func (nullLine *NullLine) setup(options string) error {
	nullLine.message = make(map[string]interface{})
	return nil
}

func (jsonLine *JsonLine) setup(options string) error {
	jsonLine.message = make(map[string]interface{})
	return nil
}

func (regexpLine *RegexpLine) setup(options string) error {
	regexpLine.message = make(map[string]interface{})
	regexp, err := regexp.Compile(options)
	if err != nil {
		return err
	}
	regexpLine.regexp = regexp
	return nil
}

func (nullLine *NullLine) parse() error {
	return nil
}

func (jsonLine *JsonLine) parse() error {
	err := json.Unmarshal(jsonLine.line, &jsonLine.message)
	if err != nil {
		return err
	}
	return nil
}

func (regexpLine *RegexpLine) parse() error {
	result := regexpLine.regexp.FindAllStringSubmatch(string(regexpLine.line), -1)
	if len(result) > 0 {
		cuttedResult := result[0]
		names := regexpLine.regexp.SubexpNames()
		for i, n := range cuttedResult {
			regexpLine.message[names[i]] = n
		}
	} else {
		errors.New("Failed to parse line using regexp: " + string(regexpLine.line))
	}
	return nil
}

func (line *Line) load(logLine []byte) {
	line.line = logLine
	line.message["Message"] = strings.TrimSpace(string(logLine))
}

func (line *Line) clear() {
	for key, _ := range line.message {
		line.message[key] = nil
	}
}

func (line *Line) addField(name string, value interface{}) {
	line.message[name] = value
}

func (line *Line) getMessage() map[string]interface{} {
	return line.message
}
