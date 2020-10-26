package apis

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func RunCommands(apiFile, outputDir string, retry *bool) bool {
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

	*retry = true
	rc := true
	output := bytes.Buffer{}
	for _, c := range cmds {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		log.Printf("Running: %s\n", c)
		cmd := exec.Command("sh", "-c", c)
		out, err := cmd.CombinedOutput()
		output.WriteString("ran: " + c + "\n")
		output.Write(out)
		output.WriteByte('\n')
		if err != nil {
			rc = false
			output.WriteString(err.Error())
			output.WriteByte('\n')
		}
		log.Printf("%s\n", string(out))
		log.Printf("err: %v\n", err)
	}
	err = ioutil.WriteFile(outputDir+"/cmds-output", output.Bytes(), 0666)
	return rc
}
