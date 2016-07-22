# Torch

_A light in dark room while you try to find a black cat._

Simple STDOUT/STDERR log forwarder & appication supervisor aimed to microservices.

**STATUS:** *experimental/alpha*


## Usage

1.Put Torchfile in working directory:

```
{
	"Service":"yaru_checker",
	"WriteHostname":true,
	"WritePort":{
		"Enabled":false,
		"Port":"12345"
	},
	"Elasticsearch":{
		"URL":"http://elasticsearch.service.consul:9200/",
		"Index":"services"
	}
}
```

2.Run torch:

`$ torch ping ya.ru`

3.Enjoy logs in realtime:

```
$ torch -l -f -n 10

```

or filter by service:

```
$ torch -l -f -s yaru_checker

```

```
$ torch -l -f -s googlecom_checker

```

_PRO TIP_

Follow all logs from all services in index:

```
$ torch -l -f -a

```

## Options

### Environment variables

All fields in Torchfile (except boolean fields) are expandable by environment variables:

```
$ export CHECKER_ID=10
$ export PORT=34534
$ export INDEX=my-index
$ export ES_ADDR=http://elasticsearch.service.consul:9200/
```

```
{
	"Service":"yaru_checker$CHECKER_ID",
	"WriteHostname":true,
	"WritePort":{
		"Enabled":true,
		"Port":"$PORT"
	},
	"Elasticsearch":{
		"URL":"$ES_ADDR",
		"Index":"$INDEX"
	}
}
```
### Log format

You can specify log format for deeper analysing by fields in high-level tools (kibana for example)

Torch can use two types of formating: `regexp` and `json`

1.`regexp` formatting for ping command:

```
{
	"Service":"yaru_checker",
	"Format": {
		"Value":"regexp",
		"Options":"(?P<bytes>\\d+) bytes from (?P<ip>(?:[0-9]{1,3}\\.){3}[0-9]{1,3}): icmp_seq=(?P<icmp_seq>\\d+) ttl=(?P<ttl>\\d+) time=(?P<time>\\d+.\\d+ [A-z]+)"
	},
	"WriteHostname":true,
	"WritePort":{
		"Enabled":false,
		"Port":"12345"
	},
	"Elasticsearch":{
		"URL":"http://elasticsearch.service.consul:9200/",
		"Index":"services"
	}
}
```

It writes additional fields `bytes`, `ip`, `icmp_seq`, `ttl` and `time` to elasticsearch.

2.`json` formating trying to unmarshal log line as JSON object:

```
{
	"Service":"yaru_checker",
	"Format": {
		"Value":"json"
	},
	"WriteHostname":true,
	"WritePort":{
		"Enabled":false,
		"Port":"12345"
	},
	"Elasticsearch":{
		"URL":"http://elasticsearch.service.consul:9200/",
		"Index":"services"
	}
}
```
