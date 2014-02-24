package logging

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/dynport/dgtk/es"
	"github.com/dynport/dgtk/util"
	"github.com/streadway/amqp"
)

var (
	allText = es.DynamicTemplates{
		{
			"all_text": es.DynamicTemplate{
				Match:            "*",
				MatchMappingType: "string",
				Mapping:          &es.DynamicTemplateMapping{Type: "string", Index: "not_analyzed"},
			},
		},
	}
)

const (
	LogsExchange = "syslog"
	DefaultTtl   = int32(60000)
)

var logger = log.New(os.Stderr, "", 0)

type Indexer struct {
	AMQPAddress        string
	ElasticSearchHost  string
	ElasticSearchIndex string
	ElasticSearchType  string
	QueueName          string
	BatchSize          int
	Ttl                int32
	Debug              bool
}

type ElasticSearchIndexMapping map[string]map[string]map[string]es.DynamicTemplates

func (indexer *Indexer) IndexMapping(indexName string) ElasticSearchIndexMapping {
	return ElasticSearchIndexMapping{
		"mappings": {indexName: {"dynamic_templates": allText}},
	}
}

func (indexer *Indexer) Run() {
	for {
		e := indexer.RunWithoutReconnect()
		if e != nil {
			fmt.Println("ERROR: " + e.Error())
			time.Sleep(1 * time.Second)
		} else {
			return
		}
	}
}

func (indexer *Indexer) NewEsIndex() *es.Index {
	return &es.Index{
		Host:      indexer.ElasticSearchHost,
		Index:     indexer.ElasticSearchIndex,
		Type:      indexer.ElasticSearchType,
		BatchSize: indexer.BatchSize,
		Debug:     indexer.Debug,
	}
}

func (indexer *Indexer) RunWithoutReconnect() error {
	con, e := amqp.Dial(indexer.AMQPAddress)
	if e != nil {
		return e
	}
	defer con.Close()
	channel, e := con.Channel()
	if e != nil {
		return e
	}
	defer channel.Close()
	t := amqp.Table{}
	if indexer.Ttl == 0 {
		indexer.Ttl = DefaultTtl
	}
	t["x-message-ttl"] = indexer.Ttl
	_, e = channel.QueueDeclare(indexer.QueueName, false, false, false, false, t)
	if e != nil {
		return e
	}
	e = channel.QueueBind(indexer.QueueName, "*", LogsExchange, false, nil)
	if e != nil {
		return e
	}
	hostname, e := os.Hostname()
	if e != nil {
		return e
	}
	consumer := hostname + ":" + strconv.Itoa(os.Getegid())
	c, e := channel.Consume(indexer.QueueName, consumer, false, false, false, false, nil)
	if e != nil {
		return e
	}
	index := indexer.NewEsIndex()
	e = indexer.CreateMappingWhenNotExists(index)
	if e != nil {
		return e
	}
	for del := range c {
		raw := string(del.Body)
		if line := parseLine(raw); line != nil {
			ok, e := index.EnqueueBulkIndex(util.MD5String(raw), line)
			if e != nil {
				logger.Print(e.Error())
			} else if ok {
				del.Ack(true)
			}
		}
	}
	index.RunBatchIndex()
	logger.Print("finished")
	return nil
}

func (indexer *Indexer) CreateMappingWhenNotExists(esIndex *es.Index) error {
	logger.Printf("getting mapping from %v", esIndex.IndexUrl())
	mapping, e := esIndex.Mapping()
	if e != nil {
		return e
	}
	logger.Printf("got mapping %#v", mapping)
	if mapping == nil {
		if esIndex.Type == "" {
			return fmt.Errorf("type must be set for elastic search index")
		}
		indexMapping := indexer.IndexMapping(esIndex.Type)
		logger.Printf("creating mapping %#v", indexMapping)
		rsp, e := esIndex.PutMapping(indexMapping)
		if e != nil {
			return e
		}
		log.Printf("create mapping status=%s", rsp.Status)
		logger.Printf("created mapping %#v", string(rsp.Body))
	} else {
		logger.Print("mapping already exists!")
	}
	return nil
}

type Parser interface {
	Parse(string) error
}

func parseLine(line string) Parser {
	parsers := []Parser{
		&NginxLine{},
		&UnicornLine{},
		&HAProxyLine{},
		&SyslogLine{},
	}
	for _, parser := range parsers {
		if e := parser.Parse(line); e == nil {
			return parser
		}
	}
	return nil
}
