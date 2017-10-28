package client

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
)

func main() {
	nodeDir := flag.String("nodeDir", "~/.gofugue", "The node package directory for custom node commands.")
	cmdString := fmt.Sprintf("/Users/wil/.nvm/versions/node/v8.6.0/bin/node %s", *nodeDir)
	fmt.Println(cmdString)
	cmd := exec.Command("/usr/bin/env", "node", "/Users/wil/.go/src/github.com/huntwj/gofugue/client/node-server") //, "/Users/wil/.gofugue")
	// cmd := exec.Command(fmt.Sprintf("/usr/bin/env node"))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(fmt.Sprintf("Could not connect to node process output.\n:: %v", err))
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(fmt.Sprintf("Could not connect to node process input.\n:: %v", err))
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(fmt.Sprintf("Error starting: %v", err))
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		buf := make([]byte, 1024)
		n, err := stdout.Read(buf)
		for ; err == nil; n, err = stdout.Read(buf) {
			fmt.Println("\n in it\n")
			test := string(buf[:n])
			fmt.Print(test)
		}
		fmt.Printf("err: %v (n: %d)\n", err, n)
		wg.Done()
	}()

	go func() {
		msg := []byte("i like it\n")
		n, err := stdin.Write(msg)
		if err != nil {
			fmt.Printf("Error writing to stdin: %v (%d)\n", err, n)
		} else {
			fmt.Printf("\nSuccessfully wrote %d bytes to stdin.\n", n)
		}
		stdin.Close()

		wg.Done()
	}()

	wg.Wait()
	os.Exit(1)
}
