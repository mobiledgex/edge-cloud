package apis

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func RunCommands(apiFile, outputDir string) bool {
	log.Printf("Running commands from %s\n", apiFile)

	if apiFile == "" {
		log.Printf("Error: cmds without API file\n")
		return false
	}

	dat, err := ioutil.ReadFile(apiFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file %s: %s\n", apiFile, err)
		return false
	}
	cmds := strings.Split(string(dat), "\n")

	rc := true
	for _, c := range cmds {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		log.Printf("Running: %s\n", c)
		cmd := exec.Command("sh", "-c", c)
		out, err := cmd.CombinedOutput()
		if err != nil {
			rc = false
		}
		log.Printf("%s\n", string(out))
		log.Printf("err: %v\n", err)
	}
	return rc
}
