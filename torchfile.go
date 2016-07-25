package main

import(
	"io"
	"os"
	"fmt"
	"log"
	"flag"
	"time"
	"bufio"
	"errors"
	"os/exec"
	"io/ioutil"
	"encoding/json"
	"gopkg.in/olivere/elastic.v3"
)

type(
	Message struct {
		Message     string    `json:"Message"`
		Service     string    `json:"Service,omitempty"`
		Timestamp   time.Time `json:"@timestamp"`
		Hostname    string    `json:"Hostname,omitempty"`
		Port        string    `json:"Port,omitempty"`
	}

	Elasticsearch struct {
		URL         string
		Index       string
	}

	WritePort struct {
		Enabled bool
		Port    string
	}

	Torchfile struct {
		Service         string
		Format          Format
		WriteHostname   bool
		WritePort       WritePort
		Elasticsearch   Elasticsearch
		client          *elastic.Client
		logChan         chan []byte
		errChan         chan error
		hostname        string
		eof             chan struct{}
	}
)

func (torchfile Torchfile) stdReader(r *bufio.Reader) {
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			torchfile.eof <- struct{}{}
			return
		}
		torchfile.logChan <- line
	}
}

func (torchfile Torchfile) exec(args []string) {
	cmd := exec.Command(args[0], args[1:len(args)]...)

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

	stdReader := io.MultiReader(stdout, stderr)
	r := bufio.NewReaderSize(stdReader, 8192)
	torchfile.eof = make(chan struct{})

	go torchfile.stdReader(r)

	log.Println("Executing application...")

	err = cmd.Start()
	if err != nil {
		log.Println("Application exited with error")
		torchfile.errChan <- err
		return
	}

	log.Println("Application is running")

	err = cmd.Wait()
	<- torchfile.eof
	time.Sleep(5 * time.Second)
	if err != nil {
		torchfile.errChan <- err
		return
	}

	torchfile.errChan <- nil
}

func (torchfile Torchfile) write() {
	var err error

	es := torchfile.Elasticsearch
	parser := torchfile.Format.parser

	for {
		parser.load(<- torchfile.logChan)
		timeNow := time.Now()
		index := es.Index + "-" + timeNow.Format("2006.01.2")

		err = parser.parse()
		if err != nil {
			log.Println(err)
		}

		parser.addField("Service", torchfile.Service)
		parser.addField("@timestamp", timeNow)
		parser.addField("Hostname", torchfile.hostname)
		parser.addField("Port", torchfile.WritePort.Port)

		_, err = torchfile.client.Index().
		Index(index).
		Type("torch").
		BodyJson(parser.getMessage()).
		Do()

		parser.clear()

		if err != nil {
			log.Println(err)
		}
	}
}

func (torchfile Torchfile) fetch(logAll bool, follow bool, followNumber int, service string) {
	var lastTimeStamp time.Time
	var message Message
	var termQuery elastic.Query

	index := torchfile.Elasticsearch.Index + "-*"

	if service == "" {
		service = torchfile.Service
	}

	if !logAll {
		termQuery = elastic.NewTermQuery("Service", service)
	} else {
		termQuery = nil
	}

	defer close(torchfile.logChan)

	searchResult, err := torchfile.client.Search().
		Index(index).
		Query(termQuery).
		Sort("@timestamp", !follow).
		From(0).Size(followNumber).
		Do()

	if err != nil {
		torchfile.errChan <- err
		return
	}

	for {
		searchResultLenght := len(searchResult.Hits.Hits)

		for i := range searchResult.Hits.Hits {
			if follow {
				i = searchResultLenght - i - 1
			}
			err := json.Unmarshal(*searchResult.Hits.Hits[i].Source, &message)
			if err != nil {
				torchfile.logChan <- []byte(err.Error())
			} else {
				torchfile.logChan <- []byte(message.Service + ": " + message.Message)
				lastTimeStamp = message.Timestamp
			}
		}


		if searchResultLenght < followNumber {
			if !follow {
				return
			}
			time.Sleep(2000 * time.Millisecond)
		} else {
			time.Sleep(500 * time.Millisecond)
		}


		rangeQuery := elastic.NewBoolQuery()
		rangeQuery = rangeQuery.Must(elastic.NewRangeQuery("@timestamp").Gt(lastTimeStamp))

		if !logAll {
			rangeQuery = rangeQuery.Filter(termQuery)
		}

		followNumber = 1024

		searchResult, err = torchfile.client.Search().
			Index(index).
			Query(rangeQuery).
			Sort("@timestamp", !follow).
			From(0).Size(followNumber).
			Do()

		if err != nil {
			torchfile.errChan <- err
			return
		}
	}
}

func (torchfile Torchfile) print() {
	var line []byte

	for line = range torchfile.logChan {
		fmt.Println(string(line))
	}

	torchfile.errChan <- nil

	return
}

func ParseTorchfile(buf []byte) (*Torchfile, error) {
	torchfile := &Torchfile{}
	var err error

	err = json.Unmarshal(buf, &torchfile)
	if err != nil {
		return nil, err
	}

	torchfile.Service = os.ExpandEnv(torchfile.Service)
	torchfile.WritePort.Port = os.ExpandEnv(torchfile.WritePort.Port)
	torchfile.Elasticsearch.URL = os.ExpandEnv(torchfile.Elasticsearch.URL)
	torchfile.Elasticsearch.Index = os.ExpandEnv(torchfile.Elasticsearch.Index)

	if torchfile.WriteHostname {
		torchfile.hostname, err = os.Hostname()
		if err != nil {
			return nil, err
		}
	}

	if !torchfile.WritePort.Enabled {
		torchfile.WritePort.Port = ""
	}


	switch torchfile.Format.Value {
	case "regexp":
		torchfile.Format.parser = &RegexpLine{}
	case "json":
		torchfile.Format.parser = &JsonLine{}
	case "":
		torchfile.Format.parser = &NullLine{}
	default:
		return nil, errors.New("No such log format: " + torchfile.Format.Value)
	}

	torchfile.Format.parser.setup(torchfile.Format.Options)

	return torchfile, nil
}

func ReadTorchfile() ([]byte, error) {
	filename := "Torchfile"
	filenameEnv := os.Getenv("TORCHFILE")
	if filenameEnv != "" {filename = filenameEnv}
	return ioutil.ReadFile(filename)
}

func (torchfile Torchfile) Run() error {
	torchfile.logChan = make(chan []byte, 1024)
	torchfile.errChan = make(chan error)

	logPtr := flag.Bool("l", false, "Show service logs")
	logAllPtr := flag.Bool("a", false, "Show all logs from all services in index")
	followPtr := flag.Bool("f", false, "Follow log updates(like 'tail -f')")
	followNumberPtr := flag.Int("n", 50, "Number of preloaded lines from log")
	servicePtr := flag.String("s", "", "Service name")
	flag.Parse()

	client, err := elastic.NewClient(elastic.SetURL(torchfile.Elasticsearch.URL))
	if err != nil {
		return err
	}

	torchfile.client = client

	if *logPtr {
		go torchfile.fetch(*logAllPtr, *followPtr, *followNumberPtr, *servicePtr)
		go torchfile.print()
	} else {
		go torchfile.exec(flag.Args())
		go torchfile.write()
	}

	return <- torchfile.errChan
}
