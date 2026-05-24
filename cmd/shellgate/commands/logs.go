package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	var follow bool
	var lines int

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "View server logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			logFile := logFilePath()

			f, err := os.Open(logFile)
			if err != nil {
				return fmt.Errorf("log file not found (%s) — has ShellGate been started yet?", logFile)
			}
			defer f.Close()

			if !follow {
				return printLastN(f, lines)
			}

			// print last N lines then follow
			if err := printLastN(f, lines); err != nil {
				return err
			}

			// seek to end, then poll for new content
			if _, err := f.Seek(0, io.SeekEnd); err != nil {
				return err
			}

			reader := bufio.NewReader(f)
			for {
				line, err := reader.ReadString('\n')
				if line != "" {
					fmt.Print(line)
				}
				if err == io.EOF {
					time.Sleep(200 * time.Millisecond)
					continue
				}
				if err != nil {
					return err
				}
			}
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow log output")
	cmd.Flags().IntVarP(&lines, "lines", "n", 50, "number of lines to show")
	return cmd
}

func printLastN(f *os.File, n int) error {
	// collect all lines, print last n
	scanner := bufio.NewScanner(f)
	var buf []string
	for scanner.Scan() {
		buf = append(buf, scanner.Text())
		if len(buf) > n {
			buf = buf[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	for _, line := range buf {
		fmt.Println(line)
	}
	return nil
}
