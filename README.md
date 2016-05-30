# Torch

_A light in dark room while you try to find a black cat._

Simple STDOUT/STDERR log forwarder & appication supervisor aimed to microservices.

**STATUS:** *experimental/alpha*


## Usage

1.Put Torchfile in working directory:

```
{
	"Command":"ping",
	"Args":["ya.ru"],
	"Service":"yaru_checker",
	"ProducerType":"elasticsearch",
	"WriteHostname":true,
	"WritePort":{
		"Enabled":false,
		"Port":12345
	},
	"Elasticsearch":{
		"URL":"http://elasticsearch.service.consul:9200/",
		"Index":"services"
	}
}
```

2.Run torch:

`$ torch`

3.Enjoy logs in elasticsearch/kibana!