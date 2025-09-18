package main

import (
	"fmt"
	"sort"
	"strings"
)

func foundAnnargams(strs []string) map[string][]string {
	anagrams := make(map[string][]string)
	for _, s := range strs {
		sCopy := strings.ToLower(s)
		sRune := []rune(sCopy)
		sort.Slice(sRune, func(i, j int) bool {
			return sRune[i] < sRune[j]
		})

		str := string(sRune)
		if _, ok := anagrams[str]; !ok {
			anagrams[str] = []string{sCopy}
		} else {
			anagrams[str] = append(anagrams[str], sCopy)
		}
	}

	ans := make(map[string][]string)

	for _, v := range anagrams {
		if len(v) == 1 {
			continue
		}
		sort.Strings(v)
		ans[v[0]] = v
	}

	return ans
}

func main() {
	input := []string{"пятак", "Пятка", "тяпка", "листок", "слиток", "столик", "стол"}
	fmt.Println(foundAnnargams(input))
}
