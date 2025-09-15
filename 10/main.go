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

var months = map[string]int{
	"Jan": 0,
	"Feb": 1,
	"Mar": 2,
	"Apr": 3,
	"May": 4,
	"Jun": 5,
	"Jul": 6,
	"Aug": 7,
	"Sep": 8,
	"Oct": 9,
	"Nov": 10,
	"Dec": 11,
}

// разделяет большой файл, чтобы сортировать по частям
const chunkSize = 1024

// получение нужного столбца
func getCol(row string, col int) string {
	cols := strings.Split(row, "\t")
	if col-1 < len(cols) {
		return cols[col-1]
	}
	return ""
}

// перевод в байты, чтобы правильно сравнивать
func parseHumanSize(s string) (float64, error) {
	multipliers := map[byte]float64{
		'K': 1024,
		'M': 1024 * 1024,
		'G': 1024 * 1024 * 1024,
	}

	n := len(s)
	if n == 0 {
		return 0, errors.New("пустая строка в определении размера")
	}
	multPart, ok := multipliers[s[n-1]]
	var numPart string
	if ok {
		numPart = s[:n-1]
	} else {
		numPart = s
		multPart = 1
	}

	val, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		return 0, err
	}
	return val * multPart, nil

}

func checkSorted(reader io.Reader, col int, flagMap map[string]bool) error {
	scanner := bufio.NewScanner(reader)
	var prev string
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		current := scanner.Text()
		if lineNum > 1 {
			rows := []string{prev, current}
			if !compareRows(0, 1, rows, col, flagMap) {
				return fmt.Errorf("нарушен порядок в строке %d: %q > %q", lineNum, prev, current)
			}
		}
		prev = current
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// объединение в один файл в отсортированном порядке
func mergeFiles(filesName []string, col int, flagMap map[string]bool) error {
	type fileRow struct {
		row   string
		index int
	}
	files := make([]*os.File, len(filesName))
	scanners := make([]*bufio.Scanner, len(filesName))
	for i, name := range filesName {
		f, err := os.Open(name)
		if err != nil {
			return err
		}
		files[i] = f
		scanners[i] = bufio.NewScanner(f)
	}
	defer func() {
		for _, f := range files {
			err := f.Close()
			if err != nil {
				fmt.Println(err)
			}
			err = os.Remove(f.Name())
			if err != nil {
				fmt.Println(err)
			}
		}
	}()
	heapRows := make([]fileRow, 0)
	for i, s := range scanners {
		if s.Scan() {
			heapRows = append(heapRows, fileRow{s.Text(), i})
		}
	}

	out := bufio.NewWriter(os.Stdout)

	for len(heapRows) > 0 {
		minIndex := 0
		minRow := heapRows[0].row
		for i, fr := range heapRows {

			if compareRows(0, 1, []string{fr.row, minRow}, col, flagMap) {
				minIndex = i
				minRow = fr.row
			}
		}
		_, err := fmt.Fprintln(out, heapRows[minIndex].row)
		if err != nil {
			return err
		}
		idx := heapRows[minIndex].index
		if scanners[idx].Scan() {
			heapRows[minIndex].row = scanners[idx].Text()
		} else {
			heapRows = append(heapRows[:minIndex], heapRows[minIndex+1:]...)
		}
	}
	err := out.Flush()
	if err != nil {
		return err
	}
	return nil
}

// компаратор для сортировки с флагами
func compareRows(i, j int, rows []string, col int, flagMap map[string]bool) bool {
	var a string
	var b string
	if col != 0 {
		a = getCol(rows[i], col)
		b = getCol(rows[j], col)
	} else {
		a = rows[i]
		b = rows[j]
	}

	if flagMap["hasB"] {
		a = strings.TrimSpace(a)
		b = strings.TrimSpace(b)
	}

	var cmp bool
	if flagMap["hasN"] {
		va, errA := strconv.ParseFloat(a, 64)
		vb, errB := strconv.ParseFloat(b, 64)
		if errA == nil && errB == nil {
			cmp = va < vb
		} else {
			cmp = a < b
		}
	} else if flagMap["hasH"] {
		va, errA := parseHumanSize(a)
		vb, errB := parseHumanSize(b)

		if errA == nil && errB == nil {
			cmp = va < vb
		} else {
			cmp = a < b
		}

	} else if flagMap["hasM"] {
		monthA, ok1 := months[a]
		monthB, ok2 := months[b]

		if ok1 && ok2 {
			cmp = monthA < monthB
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

// инициализация флагов и разбивка файла, если он слишком большой(настраивается через константу chunSize)
func procces(flags []string, in io.Reader) error {

	flagMap := make(map[string]bool)
	column := 0
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
			case 'M':
				flagMap["hasM"] = true
			case 'b':
				flagMap["hasB"] = true
			case 'c':
				flagMap["hasC"] = true
			case 'h':
				flagMap["hasH"] = true
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

	if flagMap["hasC"] {
		err := checkSorted(in, column, flagMap)
		if err != nil {
			return err
		}
		return nil
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
				_, err := fmt.Fprintln(w, s)
				if err != nil {
					return err
				}
			}
			err = w.Flush()
			if err != nil {
				return err
			}

			err = tmpFile.Close()
			if err != nil {
				return err
			}

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
			_, err := fmt.Fprintln(w, line)
			if err != nil {
				return err
			}
		}
		err = w.Flush()
		if err != nil {
			return err
		}
		err = tmpFile.Close()
		if err != nil {
			return err
		}
		chunkFiles = append(chunkFiles, tmpFile.Name())
	}

	if len(chunkFiles) == 0 {
		return nil
	}
	if len(chunkFiles) == 1 {
		f, err := os.Open(chunkFiles[0])
		if err != nil {
			return err
		}
		defer func() {
			if err := f.Close(); err != nil {
				fmt.Println(err)
			}
		}()

		_, err = io.Copy(os.Stdout, f)
		if err != nil {
			return err
		}
		err = os.Remove(chunkFiles[0])
		if err != nil {
			return err
		}
		return nil
	}

	return mergeFiles(chunkFiles, column, flagMap)
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
