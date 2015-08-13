package progress

import (
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"time"
)

var DefaultLogger Logger = log.New(os.Stderr, "", 0)

type Logger interface {
	Printf(string, ...interface{})
}

func WithTotal(total int) func(*Progress) {
	return func(p *Progress) {
		p.total = total
	}
}

type Progress struct {
	total   int
	current int
	started time.Time
	logger  Logger

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

func Start(l Logger, funcs ...func(*Progress)) *Progress {
	p := newProgress()
	p.logger = l
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
	l := p.logger
	if l == nil {
		l = DefaultLogger
	}
	t := time.Tick(dur)
	p.closer = make(chan struct{})
	p.closed = make(chan struct{})
	p.inc = make(chan int)
	go func() {
		defer close(p.closed)
		for {
			select {
			case i := <-p.inc:
				p.current += i
			case <-t:
				l.Printf("%s", p)
			case <-p.closer:
				l.Printf("%s", p)
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

func IntLen(i int) int {
	if i == 0 {
		return 1
	} else if i < 0 {
		return IntLen(int(math.Abs(float64(i)))) + 1
	}
	return int(math.Ceil(math.Log10(float64(i + 1))))
}

func (p *Progress) String() string {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	diff := p.Diff().Seconds()
	perSecond := float64(p.current) / diff
	s := fmt.Sprintf("total_time=%.06f per_second=%.01f", diff, perSecond)
	if p.total > 0 {
		l := IntLen(p.total)
		eta := float64(p.total-p.current) / perSecond
		s = fmt.Sprintf("cnt=%0*d/%0*d ", l, p.current, l, p.total) + s + fmt.Sprintf(" eta=%.01f", eta)
	} else {
		s = fmt.Sprintf("cnt=%d ", p.current) + s
	}
	return s + fmt.Sprintf(" mem_alloc=%s", SizePretty(float64(memStats.Alloc)))
}
