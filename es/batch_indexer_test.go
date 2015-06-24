package es

import "testing"

func TestBatchIndexer(t *testing.T) {
	i := NewBatchIndexer("http://127.0.0.1:9200")
	i.Close()
}
