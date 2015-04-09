package progress

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type Progress struct {
	Total         int64
	Writer        io.Writer
	Frequency     time.Duration
	Suffix        chan string
	current       int64
	c             chan int64
	finished      chan struct{}
	started       time.Time
	last          string
	currentSuffix string
	closed        bool
}

func (p *Progress) Start() chan int64 {
	p.c = make(chan int64)
	p.finished = make(chan struct{})
	p.Suffix = make(chan string)
	p.started = time.Now()
	if p.Writer == nil {
		p.Writer = os.Stdout
	}
	go func() {
		f := p.Frequency
		if f == 0 {
			f = 100 * time.Millisecond
		}
		t := time.Tick(f)
		defer func() {
			p.flush()
			fmt.Fprintf(p.Writer, "\n")
			close(p.finished)
		}()
		for {
			select {
			case <-t:
				p.flush()
			case s := <-p.Suffix:
				p.currentSuffix = s
			case i, ok := <-p.c:
				p.current += int64(i)
				if !ok {
					return
				}
			}
		}
	}()
	return p.c
}

func (p *Progress) Close() error {
	if p.closed {
		return nil
	}
	close(p.c)
	<-p.finished
	p.closed = true
	return nil
}

func (p *Progress) flush() error {
	diff := time.Since(p.started)
	perSecond := float64(p.current) / diff.Seconds()
	parts := []string{
		fmt.Sprintf("%.0f", diff.Seconds()),
		strconv.FormatInt(p.current, 10),
		fmt.Sprintf("%.03f/second", perSecond),
	}
	if p.Total > 0 {
		perc := fmt.Sprintf("%.1f%%", 100.0*float64(p.current)/float64(p.Total))
		pending := p.Total - p.current
		toGo := time.Duration(float64(pending)/perSecond) * time.Second
		parts = append(parts, perc, toGo.String())
	}
	line := fmt.Sprint(strings.Join(parts, " "))
	if p.currentSuffix != "" {
		line += " " + p.currentSuffix
	}
	clear := fmt.Sprintf("\r%s\r", strings.Repeat(" ", len(p.last)))
	p.last = line
	_, e := fmt.Fprint(p.Writer, clear+line)
	return e
}
