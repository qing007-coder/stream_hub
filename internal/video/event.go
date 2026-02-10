package video

import (
	"github.com/IBM/sarama"
	"github.com/goccy/go-json"
	"log"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/mq"
	"time"
)

type EventSender struct {
	events    []*storage.Event
	timer     *time.Timer
	duration  time.Duration
	eventChan chan *storage.Event
	maxEvents int
	producer  sarama.AsyncProducer
}

func NewEventSender(commonConf *config.CommonConfig, videoConf *config.VideoConfig) (*EventSender, error) {
	sender := new(EventSender)
	sender.events = make([]*storage.Event, 0)
	sender.duration = time.Duration(videoConf.SendEventDuration) * time.Second
	sender.maxEvents = videoConf.MaxEvent
	sender.timer = time.NewTimer(sender.duration)
	sender.eventChan = make(chan *storage.Event, 100)
	producer, err := mq.NewKafkaProducer(commonConf)
	if err != nil {
		return nil, err
	}
	sender.producer = producer.Producer()
	return sender, nil
}

func (e *EventSender) Run() {
	for {
		select {
		case <-e.timer.C:
			if err := e.flush(); err != nil {
				log.Println("flush event error: ", err)
			}
		case event := <-e.eventChan:
			if err := e.append(event); err != nil {
				log.Println("append event error: ", err)
			}
		}
	}
}

func (e *EventSender) listen() {
	for {
		select {
		case _ = <-e.producer.Successes():
		case err := <-e.producer.Errors():
			log.Println("err:", err)
		}
	}
}

func (e *EventSender) append(event *storage.Event) error {
	e.events = append(e.events, event)
	if len(e.events) >= e.maxEvents {
		return e.flush()
	}

	return nil
}

func (e *EventSender) flush() error {
	if len(e.events) == 0 {
		e.timer.Reset(e.duration)
		return nil
	}

	snapshot := e.events
	e.events = e.events[:0]

	if !e.timer.Stop() {
		<-e.timer.C
	}
	e.timer.Reset(e.duration)

	data, err := json.Marshal(&snapshot)
	if err != nil {
		return err
	}

	e.producer.Input() <- &sarama.ProducerMessage{
		Topic: constant.EventTopic,
		Value: sarama.ByteEncoder(data),
	}

	return nil
}

func (e *EventSender) Send(event *storage.Event) {
	e.eventChan <- event
}
