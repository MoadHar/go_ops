package streamz

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type Filevuep struct {
	Table  string
	View   string
	Method string
	Path   string
	Pos    int
	err    error
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	f, err := os.Open("FILES/FILEVUEP")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	// start decoding the file one line at a time
	ch := DecodeFilevuep(ctx, f)

	// read each line of output and write record to screen
	for fileview := range ch {
		if fileview.err != nil {
			continue
		}
	}
}

func DecodeFilevuep(ctx context.Context, r io.Reader) chan Filevuep {
	ch := make(chan Filevuep, 1)
	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			if ctx.Err() != nil {
				ch <- Filevuep{err: ctx.Err()}
				fmt.Println("ctx.err")
				return
			}
			v, err := GetViewpRec(scanner.Text())
			if err != nil {
				v.err = err
				fmt.Println("get view err")
				ch <- v
				return
			}
			ch <- v
		}
	}()
	fmt.Println("ok decoding...")

	return ch
}

func GetViewpRec(line string) (Filevuep, error) {
	splited := strings.Fields(line)
	if len(splited) != 5 {
		return Filevuep{err: fmt.Errorf("error: %s", line)}, nil
	} else {
		pos, err := strconv.Atoi(splited[4])
		if err != nil {
			return Filevuep{err: fmt.Errorf("error: %s", line)}, nil

		} else {
			v := Filevuep{
				Table:  strings.TrimSpace(splited[0]),
				View:   strings.TrimSpace(splited[1]),
				Method: strings.TrimSpace(splited[2]),
				Path:   strings.TrimSpace(splited[3]),
				Pos:    pos,
				err:    nil,
			}
			res := strings.TrimSpace(v.Table) == "F100"
			if res {
				fmt.Println(v.View, " -- ", v.Method, " -- ", v.Path, " -- ", v.Table)
			}
			return v, nil
		}
	}
}
