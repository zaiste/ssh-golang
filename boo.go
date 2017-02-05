package main

import (
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("ssh", "nu00", "du -s /")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
