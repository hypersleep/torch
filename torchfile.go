package main

import(
	"bufio"
	"os/exec"
	"errors"
)

type Torchfile struct {
	Command         string
	Args            []string
	Description     string
	ProducerType    string
	Elasticsearch   ElasticsearchProducer
	logChan         chan []byte
	errChan         chan error
}

func (torchfile Torchfile) Run() error {
	torchfile.logChan = make(chan []byte)
	torchfile.errChan = make(chan error)

	switch torchfile.ProducerType {
	case "elasticsearch":
		go torchfile.exec()
		go torchfile.Elasticsearch.write(torchfile.logChan, torchfile.Description)
		return <- torchfile.errChan
	default:
		return errors.New("No such producer: " + torchfile.ProducerType)
	}
	return nil
}

func (torchfile Torchfile) exec() {
	cmd := exec.Command(torchfile.Command, torchfile.Args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		torchfile.errChan <- err
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		torchfile.errChan <- err
		return
	}

	go func(){
		for {
			r := bufio.NewReader(stdout)
			line, _, err := r.ReadLine()
			if err != nil {
				torchfile.errChan <- err
				return
			}
			torchfile.logChan <- line
		}
	}()

	go func(){
		for {
			r := bufio.NewReader(stderr)
			line, _, _ := r.ReadLine()
			if err != nil {
				torchfile.errChan <- err
				return
			}
			torchfile.logChan <- line
		}
	}()

	err = cmd.Start()
	if err != nil {
		torchfile.errChan <- err
		return
	}

	err = cmd.Wait()
	if err != nil {
		torchfile.errChan <- err
		return
	}
}
