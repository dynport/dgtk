package pubsub

type Stats struct {
	received       int64
	receivedChan   chan interface{}
	dispatchedChan chan interface{}
	dispatched     int64
	collecting     bool
}

func (stats *Stats) Dispatched() int64 {
	return stats.dispatched
}

func (stats *Stats) Received() int64 {
	return stats.received
}

func (stats *Stats) MessageDispatched() {
	stats.StartCollecting()
	stats.dispatchedChan <- nil
}

func (stats *Stats) StartCollecting() {
	if stats.collecting {
		return
	}
	stats.dispatchedChan = make(chan interface{}, 1000)
	stats.receivedChan = make(chan interface{}, 1000)
	go func() {
		for {
			select {
			case <-stats.receivedChan:
				stats.received++
			case <-stats.dispatchedChan:
				stats.dispatched++
			}
		}
	}()
	stats.collecting = true
}

func (stats *Stats) MessageReceived() {
	stats.StartCollecting()
	stats.receivedChan <- nil
}

