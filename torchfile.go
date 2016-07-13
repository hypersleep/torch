package main

import(
	"os"
	"fmt"
	"log"
	"flag"
	"time"
	"bufio"
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
		WriteHostname   bool
		WritePort       WritePort
		Elasticsearch   Elasticsearch
		client          *elastic.Client
		logChan         chan []byte
		errChan         chan error
		hostname        string
	}
)

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

	log.Println("Executing application...")

	err = cmd.Start()
	if err != nil {
		log.Println("Application exited with error")
		torchfile.errChan <- err
		return
	}

	log.Println("Application is running")

	err = cmd.Wait()
	if err != nil {
		// wait for stdout and stderr
		// TODO: wait for EOF
		time.Sleep(5 * time.Second)
		torchfile.errChan <- err
		return
	}
}

func (torchfile Torchfile) write() {
	var err error
	es := torchfile.Elasticsearch

	for {
		line := string(<- torchfile.logChan)
		timeNow := time.Now()
		index := es.Index + "-" + timeNow.Format("2006.01.2")

		message := Message{
			Message: line,
			Service: torchfile.Service,
			Timestamp: timeNow,
			Hostname: torchfile.hostname,
			Port: torchfile.WritePort.Port,
		}
		_, err = torchfile.client.Index().
		Index(index).
		Type("torch").
		BodyJson(message).
		Do()

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
	if err!= nil {
		return nil, err
	}

	torchfile.Service = os.ExpandEnv(torchfile.Service)
	torchfile.WritePort.Port = os.ExpandEnv(torchfile.WritePort.Port)
	torchfile.Elasticsearch.URL = os.ExpandEnv(torchfile.Elasticsearch.URL)
	torchfile.Elasticsearch.Index = os.ExpandEnv(torchfile.Elasticsearch.Index)

	if torchfile.WriteHostname {
		torchfile.hostname, err = os.Hostname()
		if err!= nil {
			return nil, err
		}
	}

	if !torchfile.WritePort.Enabled {
		torchfile.WritePort.Port = ""
	}

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
