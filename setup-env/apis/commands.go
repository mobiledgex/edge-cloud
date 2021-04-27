package apis

import (
	"bytes"
	"encoding/csv"
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
		if strings.HasPrefix(c, "#") {
			// comment
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
	if err != nil {
		log.Printf("failed to write output, %v\n", err)
		return false
	}
	return rc
}

// Script may be a python, perl, shell, etc script.
func RunScript(apiFile, outputDir string, retry *bool) bool {
	log.Printf("Running script: %s\n", apiFile)
	*retry = true

	// api file is the command to run, not a file
	r := csv.NewReader(strings.NewReader(apiFile))
	r.Comma = ' '
	args, err := r.Read()
	if err != nil {
		log.Printf("Failed to split command into args, %v\n", err)
		return false
	}
	// Expected result can either be just a successful return value,
	// or we may also want to compare the script output.
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	log.Printf("%s\n%s\n", string(out), err)
	if err != nil {
		return false
	}
	err = ioutil.WriteFile(outputDir+"/cmds-output", out, 0666)
	if err != nil {
		log.Printf("failed to write output, %v\n", err)
		return false
	}
	return true
}
