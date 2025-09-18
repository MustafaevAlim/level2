package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type customFlags struct {
	flagF string
	flagD string
	flagS bool
}

func parseFlagF(flagF string) []string {
	return strings.Split(flagF, ",")
}

func strWithFlagF(cols []string, flagF string) (string, error) {
	var b strings.Builder
	for _, f := range parseFlagF(flagF) {
		if strings.Contains(f, "-") {
			start, err := strconv.Atoi(string(f[0]))
			if err != nil {
				return "", err
			}
			end, err := strconv.Atoi(string(f[2]))
			if err != nil {
				return "", err
			}
			if end > len(cols) {
				end = len(cols)
			}
			for i := start - 1; i < end; i++ {
				b.WriteString(cols[i])
				b.WriteString(" ")
			}

		} else {
			i, err := strconv.Atoi(f)
			if err != nil {
				return "", err
			}
			if i > len(cols) {
				continue
			}
			b.WriteString(cols[i-1])
			b.WriteString(" ")
		}

	}
	return b.String(), nil

}

func cutWithFlags(in io.Reader, flags customFlags) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		str := scanner.Text()
		cols := strings.Split(str, flags.flagD)
		if flags.flagS && len(cols) == 1 {
			continue
		}
		str, err := strWithFlagF(cols, flags.flagF)
		if err != nil {

			return err
		}
		fmt.Println(str)
	}
	return scanner.Err()
}

func process(flags customFlags, args []string) error {

	if len(args) == 0 {
		return cutWithFlags(os.Stdin, flags)
	} else {
		for _, f := range args {

			file, err := os.Open(f)
			if err != nil {
				return err
			}
			err = cutWithFlags(file, flags)
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func main() {

	flagF := flag.String("f", "1", "указание номеров полей (колонок), которые нужно вывести")
	flagD := flag.String("d", "\t", " использовать другой разделитель (символ). По умолчанию разделитель — табуляция ('\t')")
	flagS := flag.Bool("s", false, "только строки, содержащие разделитель. Если флаг указан, то строки без разделителя игнорируются")
	flag.Parse()

	flags := customFlags{
		flagF: *flagF,
		flagD: *flagD,
		flagS: *flagS,
	}

	_ = process(flags, flag.Args())

}
