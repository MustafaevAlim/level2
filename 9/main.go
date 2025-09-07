package main

import (
	"errors"
	"fmt"
	"unicode"
)

func UnpackString(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	var result []rune
	var prev rune
	havePrev := false
	escaped := false

	for _, r := range s {
		if escaped {
			prev = r
			havePrev = true
			escaped = false
			continue
		}

		if r == '\\' {
			if havePrev {
				result = append(result, prev)
			}
			havePrev = false
			escaped = true
			continue
		}

		if unicode.IsDigit(r) {
			if !havePrev {
				return "", errors.New("некорректная строка: цифра без предыдущего символа")
			}
			n := int(r - '0')
			if n > 0 {
				for i := 0; i < n; i++ {
					result = append(result, prev)
				}
			}
			havePrev = false
			continue
		}

		if havePrev {
			result = append(result, prev)
		}
		prev = r
		havePrev = true
	}

	if escaped {
		return "", errors.New("некорректная строка: оборванная escape-последовательность")
	}

	if havePrev {
		result = append(result, prev)
	}

	return string(result), nil
}
func main() {
	s, err := UnpackString("qwe\\4\\5")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(s)
}
