package es

import (
	"log"
	"os"
	"time"
)

var timer = time.NewTimer(1 * time.Second)

type IndexerStats struct {
	Runs        int64
	IndexedDocs int64
	TotalTime   time.Duration
}

func (stats *IndexerStats) Add(count int, dur time.Duration) {
	stats.Runs++
	stats.IndexedDocs += int64(count)
	stats.TotalTime += dur
}

type Indexer struct {
	Index *Index

	IndexEvery time.Duration // triggers a new index run after that duration, will be reset when BatchSize reached
	BatchSize  int           // triggers a new index run when the batch reaches that size

	docsBatch   []*Doc
	docsChannel chan *Doc
	timer       *time.Timer
	Stats       IndexerStats
}

func (indexer *Indexer) Finish() error {
	return indexer.indexBatch()
}

func (indexer *Indexer) resetBatch() {
	indexer.docsBatch = make([]*Doc, 0, indexer.BatchSize)
}

func (indexer *Indexer) resetTimer() bool {
	return indexer.timer.Reset(indexer.IndexEvery)
}

var logger = log.New(os.Stderr, "", 0)

// TODO change handling to e.g. Add command with implizit opening of channel
// to not expose the channel to the public
func (indexer *Indexer) Start() chan *Doc {
	if indexer.BatchSize == 0 {
		indexer.BatchSize = 100
	}
	if indexer.IndexEvery == 0 {
		indexer.IndexEvery = 1 * time.Hour
	}
	indexer.timer = time.NewTimer(indexer.IndexEvery)
	indexer.docsChannel = make(chan *Doc, indexer.BatchSize)
	indexer.resetBatch()
	go func(indexer *Indexer) {
		for {
			select {
			case <-indexer.timer.C:
				e := indexer.indexBatch()
				if e != nil {
					logger.Printf("ERROR=%q", e)
				}
				indexer.resetTimer()
			case doc, ok := <-indexer.docsChannel:
				if !ok {
					indexer.timer.Stop()
					e := indexer.indexBatch()
					if e != nil {
						logger.Printf("ERROR=%q", e)
					}
					return
				}
				if doc.Index == "" {
					doc.Index = indexer.Index.Index
				}
				if doc.Type == "" {
					doc.Type = indexer.Index.Type
				}
				indexer.docsBatch = append(indexer.docsBatch, doc)
				if len(indexer.docsBatch) >= indexer.BatchSize {
					indexer.timer.Stop()
					indexer.indexBatch()
					indexer.resetTimer()
				}
			}
		}
	}(indexer)
	return indexer.docsChannel
}

func (indexer *Indexer) Close() error {
	close(indexer.docsChannel)
	return indexer.indexBatch()
}

func (indexer *Indexer) indexBatch() error {
	if len(indexer.docsBatch) < 1 {
		return nil
	}
	started := time.Now()
	e := indexer.Index.IndexDocs(indexer.docsBatch)
	if e != nil {
		return e
	}
	indexer.Stats.Add(len(indexer.docsBatch), time.Now().Sub(started))
	indexer.resetBatch()
	return nil
}
