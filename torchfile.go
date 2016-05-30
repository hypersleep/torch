package main

import(
	"time"
	"bufio"
	"os/exec"
	"errors"
)

type(
	WritePort struct {
		Enabled bool
		Port    int
	}

	Torchfile struct {
		Command         string
		Args            []string
		Service         string
		ProducerType    string
		WriteHostname   bool
		WritePort       WritePort
		Elasticsearch   ElasticsearchProducer
		logChan         chan []byte
		errChan         chan error
		hostname        string
	}
)

func (torchfile Torchfile) Run() error {
	torchfile.logChan = make(chan []byte, 1024)
	torchfile.errChan = make(chan error)

	switch torchfile.ProducerType {
	case "elasticsearch":
		go torchfile.exec()
		go torchfile.Elasticsearch.write(torchfile.logChan, torchfile.Service, torchfile.hostname, torchfile.WritePort.Port)
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
			line, _, _ := r.ReadLine()
			torchfile.logChan <- line
		}
	}()

	go func(){
		for {
			r := bufio.NewReader(stderr)
			line, _, _ := r.ReadLine()
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
		// wait for stdout and stderr
		time.Sleep(5 * time.Second)
		torchfile.errChan <- err
		return
	}
}

// func (torchfile *Torchfile) SetHostname() {
// 	torchfile.hostname
// }
