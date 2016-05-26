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
		log.Fatal("Failed to read Torchfile:", err)
	}

	torchfile := &Torchfile{}
	err = json.Unmarshal(buf, &torchfile)
	if err!= nil {
		log.Fatal("Failed to unmarshal torchfile:", err)
	}

	torchfile.Description = os.ExpandEnv(torchfile.Description)

	err = torchfile.Run()
	if err != nil {
		log.Fatal("Producer error:", err)
	}
}