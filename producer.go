package main

import(
	"gopkg.in/olivere/elastic.v3"
	"log"
	"time"
)

type Producer interface {
	write() error
}

type ElasticsearchProducer struct {
	URL         string
	Index       string
}

type ElasticsearchMessage struct {
	Message     string    `json:"Message"`
	Service string    `json:"Service"`
	Timestamp   time.Time `json:"@timestamp"`
}

func (esProducer ElasticsearchProducer) write(logChan chan []byte, service string) {
	client, err := elastic.NewClient(elastic.SetURL(esProducer.URL))
	if err != nil {
		log.Println(err)
	}

	_, err = client.CreateIndex(esProducer.Index).Do()
	if err != nil {
		log.Println(err)
	}

	for {
		line := string(<- logChan)
		timeNow := time.Now()
		index := esProducer.Index + "-" + timeNow.Format("2006-01-2")

		message := ElasticsearchMessage{
			Message: line,
			Service: service,
			Timestamp: timeNow,
		}
		_, err = client.Index().
		Index(index).
		Type("torch").
		BodyJson(message).
		Do()

		if err != nil {
			log.Println(err)
		}
	}
	return
}
