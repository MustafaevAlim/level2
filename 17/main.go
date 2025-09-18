package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Session struct {
	Conn   net.Conn
	In     chan string
	Out    chan string
	ErrCh  chan error
	Ctx    context.Context
	Cancel context.CancelFunc
	Wg     *sync.WaitGroup
	M      *sync.Mutex
}

func (s *Session) ReadConn() {
	defer s.Wg.Done()
	defer close(s.Out)
	scanner := bufio.NewScanner(s.Conn)
	for {
		select {
		case <-s.Ctx.Done():
			return
		default:
			if scanner.Scan() {
				select {
				case s.Out <- scanner.Text():
				case <-s.Ctx.Done():
					return
				}
			} else {
				return
			}
		}
	}
}

func (s *Session) WriteConn() {
	defer s.Wg.Done()
	writer := bufio.NewWriter(s.Conn)
	for {
		select {
		case <-s.Ctx.Done():
			return
		case in, ok := <-s.In:
			if !ok {
				return
			}
			s.M.Lock()
			fmt.Println(in)
			_, err := writer.Write([]byte(in))
			if err != nil {
				log.Println(err)
			}
			if err := writer.Flush(); err != nil {
				log.Println(err)
			}
			s.M.Unlock()
		case err, ok := <-s.ErrCh:
			if !ok {
				return
			}
			if errors.Is(err, io.EOF) {
				s.Cancel()
				return
			}
			log.Println("Ошибка записи:", err)
		}
	}
}

func (s *Session) ReadInput() {
	defer close(s.In)
	defer close(s.ErrCh)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-s.Ctx.Done():
			return
		default:
			if scanner.Scan() {
				s.In <- scanner.Text()
			} else if err := scanner.Err(); err != nil {
				s.ErrCh <- err
			} else {
				s.ErrCh <- io.EOF
				return
			}
		}
	}
}

func main() {
	var m sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)
	host := flag.String("host", "", "- host server")
	port := flag.String("port", "", "- port server")
	timeout := flag.Int("timeout", 0, "- timeout connection")
	flag.Parse()

	conn, err := net.DialTimeout("tcp", *host+":"+*port, time.Duration(*timeout)*time.Second)
	if err != nil {
		log.Println(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	in := make(chan string, 10)
	out := make(chan string, 10)
	errCh := make(chan error, 1)

	session := Session{
		Conn:   conn,
		In:     in,
		Out:    out,
		ErrCh:  errCh,
		Wg:     &wg,
		M:      &m,
		Ctx:    ctx,
		Cancel: cancel,
	}

	go session.ReadInput()
	go session.ReadConn()
	go session.WriteConn()
loop:
	for {
		select {
		case str := <-out:
			fmt.Println(str)
		case <-ctx.Done():
			err := conn.Close()
			if err != nil {
				log.Println(err)
			}
			fmt.Println("Завершение соединения")
			break loop
		}
	}
	cancel()

	wg.Wait()

}
