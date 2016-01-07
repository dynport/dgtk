package progress

import (
	"fmt"
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

func HumanReadable(p *Progress) {
	p.humanReadable = true
}

func WithPrinter(printer Printer) StartFunc {
	return func(p *Progress) {
		p.printer = printer
	}
}

type Progress struct {
	total         int
	humanReadable bool
	current       int
	started       time.Time
	printer       Printer

	inc      chan int
	closer   chan struct{}
	closed   chan struct{}
	isClosed bool
}

func newProgress() *Progress {
	return &Progress{started: time.Now().UTC()}
}

func (p *Progress) Write(b []byte) (int, error) {
	p.inc <- len(b)
	return len(b), nil
}

type LogPrinter struct {
	HumanReadable bool
	Logger        Logger
}

func (p *LogPrinter) Print(s *Status) {
	p.Logger.Printf(statusToString(s, p.HumanReadable))
}

func statusToString(status *Status, humanReadable bool) string {
	perSecond := func() string {
		if humanReadable {
			return SizePretty(status.PerSecond())
		} else {
			return fmt.Sprintf("%.01f", status.PerSecond())
		}
	}()
	s := fmt.Sprintf("total_time=%.06f per_second=%s", status.RunningSince().Seconds(), perSecond)
	if status.Total != nil {
		l := IntLen(*status.Total)
		prefix := func() string {
			if humanReadable {
				return SizePretty(float64(status.Current)) + "/" + SizePretty(float64(*status.Total))
			}
			return fmt.Sprintf("cnt=%0*d/%0*d", l, status.Current, l, *status.Total)
		}()
		s = prefix + " " + s + fmt.Sprintf(" eta=%.01f", status.ETA().Seconds())
	} else {
		s = fmt.Sprintf("cnt=%d ", status.Current) + s
	}
	if status.MemStats.Alloc > 0 {
		s += fmt.Sprintf(" mem_alloc=%s", SizePretty(float64(status.MemStats.Alloc)))
	}
	return s
}

func Start(l Logger, funcs ...func(*Progress)) *Progress {
	p := newProgress()
	for _, f := range funcs {
		f(p)
	}
	p.printer = &LogPrinter{Logger: l, HumanReadable: p.humanReadable}
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
