package main

import (
	"fmt"
  "os/exec"
)
func main() {
	cmd := exec.Command("ansible", "demo", "-m", "ping")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("cmd.Run() failed with %s\n", err)
	}
	fmt.Printf("combined out:\n%s\n", string(out))
}
