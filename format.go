package main

import(
	"regexp"
	"errors"
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
		load(line []byte)
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

func (nullLine *NullLine) load(line []byte) {
	nullLine.line = line
	nullLine.message["Message"] = string(line)
}

func (jsonLine *JsonLine) load(line []byte) {
	jsonLine.line = line
	jsonLine.message["Message"] = string(line)
}

func (regexpLine *RegexpLine) load(line []byte) {
	regexpLine.line = line
	regexpLine.message["Message"] = string(line)
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

func (nullLine *NullLine) clear() {
	for key, _ := range nullLine.message {
		nullLine.message[key] = nil
	}
}

func (jsonLine *JsonLine) clear() {
	for key, _ := range jsonLine.message {
		jsonLine.message[key] = nil
	}
}

func (regexpLine *RegexpLine) clear() {
	for key, _ := range regexpLine.message {
		regexpLine.message[key] = nil
	}
}

func (nullLine *NullLine) addField(name string, value interface{}) {
	nullLine.message[name] = value
}

func (jsonLine *JsonLine) addField(name string, value interface{}) {
	jsonLine.message[name] = value
}

func (regexpLine *RegexpLine) addField(name string, value interface{}) {
	regexpLine.message[name] = value
}

func (nullLine *NullLine) getMessage() map[string]interface{} {
	return nullLine.message
}

func (jsonLine *JsonLine) getMessage() map[string]interface{} {
	return jsonLine.message
}

func (regexpLine *RegexpLine) getMessage() map[string]interface{} {
	return regexpLine.message
}
