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
