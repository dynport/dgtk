package progress

import (
	"fmt"
	"time"
)

type Writer struct {
	total   int64
	current int64
	c       chan int
	closed  chan struct{}
	doClose chan struct{}
}

func NewWriter() *Writer {
	return NewWriterWithTotal(0)
}

func NewWriterWithTotal(total int64) *Writer {
	p := &Writer{
		total:   total,
		c:       make(chan int),
		doClose: make(chan struct{}),
		closed:  make(chan struct{}),
	}
	go p.start()
	return p
}

func (p *Writer) start() {
	defer close(p.closed)
	started := time.Now()
	t := time.Tick(1 * time.Second)
	// declared here so it can not be used outside
	doPrint := func() {
		diff := time.Since(started).Seconds()
		perSecond := float64(p.current) / diff
		out := fmt.Sprintf("%s/second %s", SizePretty(perSecond), SizePretty(float64(p.current)))
		if p.total > 0 {
			out += "/" + SizePretty(float64(p.total))
			pending := p.total - p.current
			timeToGo := time.Duration(int(float64(pending)/perSecond)) * time.Second
			out += " " + fmt.Sprintf("eta=%s", timeToGo.String())
		}
		fmt.Print(out + "\n")
	}
	for {
		select {
		case i := <-p.c:
			p.current += int64(i)
		case <-t:
			doPrint()
		case <-p.doClose:
			doPrint()
			return
		}
	}
}

func (p *Writer) Close() error {
	close(p.doClose)
	<-p.closed
	return nil
}

func (w *Writer) Write(b []byte) (int, error) {
	w.c <- len(b)
	return len(b), nil

}
