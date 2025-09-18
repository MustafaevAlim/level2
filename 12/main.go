package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

type flags struct {
	flagA         int
	flagB         int
	flagC         bool
	flagI         bool
	flagV         bool
	flagF         bool
	flagN         bool
	printFilename bool
}

func foundStringWithFlags(in io.Reader, param flags, substr string, filename string) error {

	scanner := bufio.NewScanner(in)
	countFound := 0
	lines := 1
	prevLines := make([]string, 0, param.flagB)
	nextLinesCount := 0
	for scanner.Scan() {
		str := scanner.Text()
		if nextLinesCount > 0 {
			fmt.Println(str)
			nextLinesCount--
		}

		match, err := isMatch(str, substr, param)
		if err != nil {
			return err
		}
		if match {
			countFound++
			if param.flagA > 0 && !param.flagC {
				nextLinesCount = param.flagA
			}

			if !param.flagC {
				showTextContext(prevLines, param)
				printLine(str, lines, param, filename)
			}

		}

		updateContext(&prevLines, str, param.flagB)
		lines++
	}

	if param.flagC {
		fmt.Println(countFound)
	}

	return scanner.Err()
}

func printLine(str string, line int, param flags, filename string) {
	b := strings.Builder{}
	if param.printFilename {
		fmt.Fprintf(&b, "%s: ", filename)
	}
	if param.flagN {
		fmt.Fprintf(&b, "%d: ", line)
	}
	fmt.Fprintf(&b, "%s", str)
	fmt.Println(b.String())

}

func updateContext(prevLines *[]string, str string, contextSize int) {
	if contextSize == 0 {
		return
	}

	if len(*prevLines) == contextSize {
		*prevLines = (*prevLines)[1:]
	}
	*prevLines = append(*prevLines, str)
}

func showTextContext(prevLines []string, param flags) {
	if param.flagB > 0 {
		for _, s := range prevLines {
			fmt.Println(s)
		}
	}
}

func isMatch(str string, substr string, param flags) (bool, error) {
	if param.flagF {
		if param.flagI {
			return strings.Contains(strings.ToLower(str), strings.ToLower(substr)), nil
		}
		return strings.Contains(str, substr), nil
	}

	if param.flagI {
		substr = "(?i)" + substr
	}
	re, err := regexp.Compile(substr)
	if err != nil {
		return false, err
	}
	return re.MatchString(str), nil
}

func procces(param flags, args []string) error {
	substr := args[0]

	if len(args) == 1 {
		return foundStringWithFlags(os.Stdin, param, substr, "")
	} else if len(args) == 0 {
		return errors.New("используйте: grep [Параметры] ... ШАБЛОНЫ [Файл]")
	} else if len(args) == 2 {
		file, err := os.Open(args[1])
		if err != nil {
			return err
		}
		err = foundStringWithFlags(file, param, substr, args[1])
		if err != nil {
			return err
		}
	} else {
		param.printFilename = true
		var wg sync.WaitGroup
		wg.Add(len(args[1:]))

		for _, f := range args[1:] {
			go func(filename string) {
				defer wg.Done()
				file, err := os.Open(filename)
				if err != nil {
					log.Println("Не удалось открыть файл", filename, ":", err)
				}
				err = foundStringWithFlags(file, param, substr, filename)
				if err != nil {
					log.Println("Ошибка при обработке файла", filename, ":", err)
				}
			}(f)
		}
		wg.Wait()
	}

	return nil
}

func main() {
	flagA := flag.Int("A", 0, "-A N — после каждой найденной строки дополнительно вывести N строк после неё")
	flagB := flag.Int("B", 0, "-B N — вывести N строк до каждой найденной строки")
	flagAB := flag.Int("C", 0, "-C N — вывести N строк контекста вокруг найденной строки")
	flagC := flag.Bool("c", false, "-c — выводить только то количество строк, что совпадающих с шаблоном")
	flagI := flag.Bool("i", false, "-i — игнорировать регистр")
	flagV := flag.Bool("v", false, "-v — инвертировать фильтр: выводить строки, не содержащие шаблон")
	flagF := flag.Bool("F", false, "-F — воспринимать шаблон как фиксированную строку, а не регулярное выражение")
	flagN := flag.Bool("n", false, "-n — выводить номер строки перед каждой найденной строкой")

	flag.Parse()

	if *flagAB > 0 {
		*flagA = *flagAB
		*flagB = *flagAB
	}

	param := flags{
		flagA: *flagA,
		flagB: *flagB,
		flagC: *flagC,
		flagI: *flagI,
		flagV: *flagV,
		flagF: *flagF,
		flagN: *flagN,
	}

	err := procces(param, flag.Args())
	if err != nil {
		fmt.Println(err)
	}

}
