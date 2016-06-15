package main

import(
	"log"
)

func main() {
	buf, err := ReadTorchfile()
	if err != nil {
		log.Fatal("Failed to read Torchfile: ", err)
	}

	torchfile, err := ParseTorchfile(buf)
	if err != nil {
		log.Fatal("Failed to parse or expand Torchfile: ", err)
	}

	err = torchfile.Run()
	if err != nil {
		log.Fatal("Torch error: ", err)
	}
}