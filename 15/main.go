package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"syscall"
)

func parseRedirect(openedFile *[]*os.File, s []string) (*exec.Cmd, error) {
	args := make([]string, 0)
	var inputFile, outputFile string
	for i := 0; i < len(s); i++ {
		if s[i] == "<" && i+1 < len(s) {
			inputFile = s[i+1]
			i++
		} else if s[i] == ">" && i+1 < len(s) {
			outputFile = s[i+1]
			i++
		} else {
			args = append(args, s[i])
		}
	}

	cmd := exec.Command(args[0], args[1:]...)
	if inputFile != "" {
		file, err := os.Open(inputFile)
		if err != nil {
			return nil, err
		}
		cmd.Stdin = file
		*openedFile = append(*openedFile, file)
	}
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return nil, err
		}
		cmd.Stdout = file
		*openedFile = append(*openedFile, file)
	}
	return cmd, nil
}

func runPipeLine(commandLine string) error {
	stages := strings.Split(commandLine, "|")

	var cmds []*exec.Cmd
	var openedFile []*os.File
	for _, stage := range stages {
		s := strings.Fields(strings.TrimSpace(stage))
		if len(s) == 0 {
			continue
		}
		if slices.Contains(s, ">") || slices.Contains(s, "<") {
			cmd, err := parseRedirect(&openedFile, s)
			if err != nil {
				return err
			}
			cmds = append(cmds, cmd)
		} else {
			cmds = append(cmds, exec.Command(s[0], s[1:]...))
		}

	}

	if len(cmds) == 0 {
		return nil
	}

	for i := 0; i < len(cmds)-1; i++ {
		if cmds[i+1].Stdin == nil && cmds[i].Stdout == nil {
			stdout, err := cmds[i].StdoutPipe()
			if err != nil {
				return err
			}
			cmds[i+1].Stdin = stdout
		}

	}
	if len(cmds) == 1 {

		if cmds[0].Stdin == nil {
			cmds[0].Stdin = os.Stdin

		}

	}
	if cmds[len(cmds)-1].Stdout == nil {
		cmds[len(cmds)-1].Stdout = os.Stdout

	}

	for _, c := range cmds {
		c.Stderr = os.Stderr
	}

	for _, c := range cmds {
		if err := c.Start(); err != nil {
			return err
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	done := make(chan error, 1)
	go func() {
		for _, c := range cmds {
			if err := c.Wait(); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()

	select {
	case sig := <-sigChan:
		for _, c := range cmds {
			if c.Process != nil {
				err := c.Process.Signal(sig)
				if err != nil {
					return err
				}
			}
		}

		for _, f := range openedFile {
			err := f.Close()
			if err != nil {
				fmt.Println(err)
			}
		}
		return errors.New("signal: interrupt")

	case err := <-done:
		for _, f := range openedFile {
			err := f.Close()
			if err != nil {
				fmt.Println(err)
			}
		}
		return err

	}

}

func processShell() {
	reader := bufio.NewReader(os.Stdin)
	for {
		currDir, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}
		currDirParts := strings.Split(currDir, "/")
		fmt.Printf("%s -> ", currDirParts[len(currDirParts)-1])
		line, _, err := reader.ReadLine()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			fmt.Println(err)
		}

		commands := strings.Fields(string(line))

		if len(commands) == 0 {
			continue
		}

		switch commands[0] {
		case "cd":
			err := os.Chdir(commands[1])
			if err != nil {
				fmt.Println(err)
			}
		case "pwd":
			fmt.Println(currDir)
		case "echo":
			fmt.Println(strings.Join(commands[1:], " "))
		case "kill":
			pid, err := strconv.Atoi(commands[1])
			if err != nil {
				fmt.Println(err)
			}
			err = syscall.Kill(pid, syscall.SIGTERM)
			if err != nil {
				fmt.Println(err)
			}
		default:

			err := runPipeLine(string(line))
			if err != nil {
				fmt.Println(err)
			}
		}

	}
}

func main() {
	processShell()
}
