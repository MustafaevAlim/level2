package main

import (
	"fmt"
	"os"
	"time"

	"github.com/beevik/ntp"
)

func GetCurrentTime() int {
	t, err := ntp.Time("0.beevik-ntp.pool.ntp.org")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ошибка: %v\n", err)
		return 1
	}

	fmt.Println(t.Format(time.RFC3339))
	return 0
}

func main() {
	os.Exit(GetCurrentTime())

}
