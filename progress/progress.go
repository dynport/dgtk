package progress

import (
	"log"
	"os"
	"runtime"
	"time"
)

var DefaultLogger Logger = log.New(os.Stderr, "", 0)

type Logger interface {
	Printf(string, ...interface{})
}

type StartFunc func(*Progress)

func WithTotal(total int) StartFunc {
	return func(p *Progress) {
		p.total = total
	}
}

func WithPrinter(printer Printer) StartFunc {
	return func(p *Progress) {
		p.printer = printer
	}
}

type Progress struct {
	total   int
	current int
	started time.Time
	printer Printer

	inc      chan int
	closer   chan struct{}
	closed   chan struct{}
	isClosed bool
}

func newProgress() *Progress {
	return &Progress{started: time.Now().UTC()}
}

func newWithTotal(total int) *Progress {
	p := newProgress()
	p.total = total
	return p
}

func (p *Progress) Write(b []byte) (int, error) {
	p.inc <- len(b)
	return len(b), nil
}

type LogPrinter struct {
	Logger Logger
}

func (p *LogPrinter) Print(s *Status) {
	p.Logger.Printf(s.String())
}

func Start(l Logger, funcs ...func(*Progress)) *Progress {
	p := newProgress()
	p.printer = &LogPrinter{l}
	for _, f := range funcs {
		f(p)
	}
	p.Start()
	return p
}

func (p *Progress) Inc() {
	p.inc <- 1
}

func (p *Progress) IncBy(i int) {
	p.inc <- i
}

func (p *Progress) Reset() {
	p.total = 0
	p.current = 0
	p.started = time.Now().UTC()
	p.isClosed = false
}

func (p *Progress) Diff() time.Duration {
	return time.Since(p.started)
}

func (p *Progress) Start() {
	dur := 1 * time.Second
	printer := p.printer
	if printer == nil {
		printer = &LogPrinter{Logger: DefaultLogger}
	}
	t := time.Tick(dur)
	p.closer = make(chan struct{})
	p.closed = make(chan struct{})
	p.inc = make(chan int)
	go func() {
		defer close(p.closed)
		printedMax := false
		for {
			select {
			case i := <-p.inc:
				p.current += i
			case <-t:
				if !printedMax {
					printer.Print(p.Status())
				}
				printedMax = p.total > 0 && p.current >= p.total
			case <-p.closer:
				printer.Print(p.Status())
				return
			}
		}
	}()
	return
}

func (p *Progress) Close() error {
	if p.isClosed {
		return nil
	}
	p.closer <- struct{}{}
	<-p.closed
	p.isClosed = true
	return nil
}

func (p *Progress) Status() *Status {
	s := &Status{
		Current: p.current,
		Started: p.started,
		Now:     time.Now(),
	}
	runtime.ReadMemStats(&s.MemStats)
	if p.total > 0 {
		s.Total = &p.total
	}
	return s
}
