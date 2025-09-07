package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const chunkSize = 10

func getCol(row string, col int) string {
	cols := strings.Split(row, "\t")
	if col-1 < len(cols) {
		return cols[col-1]
	}
	return ""
}

func mergeFiles(filesName []string) error {

}

func compareRows(i, j int, rows []string, col int, flagMap map[string]bool) bool {

	a := getCol(rows[i], col)
	b := getCol(rows[j], col)

	var cmp bool
	if flagMap["hasN"] {
		va, errA := strconv.ParseFloat(a, 64)
		vb, errB := strconv.ParseFloat(b, 64)
		if errA == nil && errB == nil {
			cmp = va < vb
		} else {
			cmp = a < b
		}
	} else {
		cmp = a < b
	}
	if flagMap["hasR"] {
		return !cmp
	}
	return cmp

}

func sortWithFlags(rows []string, col int, flagMap map[string]bool) []string {
	sort.Slice(rows, func(i, j int) bool {
		return compareRows(i, j, rows, col, flagMap)
	})

	if flagMap["hasU"] {
		uniqRowsMap := make(map[string]bool)
		uniqRows := make([]string, 0, len(rows))
		for _, r := range rows {
			uniqRowsMap[r] = true
		}
		for k := range uniqRowsMap {
			uniqRows = append(uniqRows, k)
		}
		rows = uniqRows

	}
	return rows
}

func printResult(strs []string) {
	for _, s := range strs {
		fmt.Println(s)
	}
}

func procces(flags []string, in io.Reader) error {

	flagMap := make(map[string]bool)
	column := 1
	var err error
	if len(flags) != 0 {
		for i := 1; i < len(flags[0]); i++ {

			switch flags[0][i] {
			case 'n':
				flagMap["hasN"] = true
			case 'r':
				flagMap["hasR"] = true
			case 'u':
				flagMap["hasU"] = true
			case 'k':
				if len(flags) < 2 {
					return errors.New("ошибка: неверный аругмент -k")
				}
				column, err = strconv.Atoi(flags[1])
				if err != nil {
					return err
				}
			}
		}
	}

	scanner := bufio.NewScanner(in)
	chunkFiles := make([]string, 0)
	strs := make([]string, 0, chunkSize)

	for scanner.Scan() {
		strs = append(strs, scanner.Text())
		if len(strs) >= chunkSize {
			strs = sortWithFlags(strs, column, flagMap)

			tmpFile, err := os.CreateTemp("", "chunk")
			if err != nil {
				return err
			}

			w := bufio.NewWriter(tmpFile)
			for _, s := range strs {
				fmt.Fprintln(w, s)
			}
			w.Flush()
			tmpFile.Close()

			chunkFiles = append(chunkFiles, tmpFile.Name())
			strs = make([]string, 0, chunkSize)
		}

	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(strs) > 0 {
		strs = sortWithFlags(strs, column, flagMap)
		tmpFile, err := os.CreateTemp("", "chunk")
		if err != nil {
			return err
		}
		w := bufio.NewWriter(tmpFile)
		for _, line := range strs {
			fmt.Fprintln(w, line)
		}
		w.Flush()
		tmpFile.Close()
		chunkFiles = append(chunkFiles, tmpFile.Name())
	}

	fmt.Println(chunkFiles)

	if len(chunkFiles) == 0 {
		return nil
	}
	if len(chunkFiles) == 1 {
		f, err := os.Open(chunkFiles[0])
		if err != nil {
			return err
		}
		defer f.Close()
		io.Copy(os.Stdout, f)
		os.Remove(chunkFiles[0])
		return nil
	}

	printResult(strs)
	return nil
}

func parseArgs(args []string) ([]string, io.Reader, error) {
	flags := make([]string, 0, 2)
	var in io.ReadCloser
	for _, v := range args[1:] {
		if v[0] == '-' {
			flags = append(flags, v)
		} else if unicode.IsDigit(rune(v[0])) {
			if len(flags) == 0 {
				return nil, nil, errors.New("неверные аргументы, используйте: sort [параметры] [файл]")
			}
			flags = append(flags, v)
		} else {
			file, err := os.Open(v)
			if err != nil {
				return nil, nil, err
			}
			in = file
		}
	}

	if in == nil {
		in = os.Stdin
	}

	return flags, in, nil

}

// как в unix sort если с флагом -n используются строки, которые нельзя интерпретировать как число, то сортируется по символам
func main() {
	flags, in, err := parseArgs(os.Args)
	if err != nil {
		fmt.Println(err)
		return
	}
	if closer, ok := in.(io.Closer); ok && in != os.Stdin {
		defer func() {
			if err := closer.Close(); err != nil {
				fmt.Println(err)
			}
		}()
	}

	err = procces(flags, in)
	if err != nil {
		fmt.Println(err)
		return
	}

}
