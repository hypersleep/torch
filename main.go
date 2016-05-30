package main

import(
	"os"
	"log"
	"io/ioutil"
	"encoding/json"
)

func main() {
	filename := "Torchfile"
	filenameEnv := os.Getenv("TORCHFILE")

	if filenameEnv != "" {filename = filenameEnv}

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Failed to read Torchfile: ", err)
	}

	torchfile := &Torchfile{}
	err = json.Unmarshal(buf, &torchfile)
	if err!= nil {
		log.Fatal("Failed to unmarshal torchfile: ", err)
	}

	torchfile.Service = os.ExpandEnv(torchfile.Service)

	for argIndex := range torchfile.Args {
		torchfile.Args[argIndex] = os.ExpandEnv(torchfile.Args[argIndex])
	}

	if torchfile.WriteHostname {
		torchfile.hostname, err = os.Hostname()
		if err!= nil {
			log.Fatal("Failed to set hostname: ", err)
		}
	}

	if !torchfile.WritePort.Enabled {
		torchfile.WritePort.Port = 0
	}

	err = torchfile.Run()
	if err != nil {
		log.Fatal("Torch error: ", err)
	}
}