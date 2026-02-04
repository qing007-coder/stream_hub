package infra

import (
	"bytes"
	"github.com/IBM/sarama"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/mq"
	"time"
)

type Logger struct {
	batchLogs   [][]byte
	producer    sarama.AsyncProducer
	logger      *zap.Logger
	maxBatchNum int
	duration    time.Duration
	timer       *time.Timer
	inputChan   chan []byte
}

func NewLogger(conf *config.CommonConfig) (*Logger, error) {
	l := new(Logger)
	producer, err := mq.NewKafkaProducer(conf)
	if err != nil {
		return nil, err
	}
	l.producer = producer.Producer()
	l.batchLogs = make([][]byte, 0)
	l.maxBatchNum = conf.Logger.MaxSize
	l.duration = time.Duration(conf.Logger.Duration) * time.Second
	l.timer = time.NewTimer(l.duration)
	l.inputChan = make(chan []byte, 50)
	l.init()

	go l.Run()
	return l, nil
}

func (l *Logger) init() {
	logMode := zapcore.DebugLevel

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(l),
		logMode,
	)
	l.logger = zap.New(core)
}

func (l *Logger) Run() {
	go l.listen()
	for {
		select {
		case <-l.timer.C:
			if err := l.flush(); err != nil {
				log.Println("err:", err)
			}

		case data := <-l.inputChan:
			l.batchLogs = append(l.batchLogs, data)
			if len(l.batchLogs) >= l.maxBatchNum {
				if err := l.flush(); err != nil {
					log.Println("err:", err)
				}
			}
		}
	}
}

func (l *Logger) listen() {
	for {
		select {
		case _ = <-l.producer.Successes():
		case err := <-l.producer.Errors():
			log.Println("err:", err)
		}
	}
}

func (l *Logger) Write(data []byte) (n int, err error) {
	select {
	case l.inputChan <- data:
	default:
		// 丢日志
	}

	return len(data), nil
}

func (l *Logger) flush() error {
	if len(l.batchLogs) == 0 {
		l.timer.Reset(l.duration) // 即使没数据，也要重置计时器
		return nil
	}

	snapshot := l.batchLogs
	l.batchLogs = make([][]byte, 0)

	if !l.timer.Stop() {
		<-l.timer.C
	}
	l.timer.Reset(l.duration) // 重置计时器

	// 拼接成 JSON 数组: [{},{}]
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, b := range snapshot {
		buf.Write(b)
		if i < len(snapshot)-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte(']')

	l.producer.Input() <- &sarama.ProducerMessage{
		Topic: constant.UserLogTopic,
		Value: sarama.ByteEncoder(buf.Bytes()),
	}
	return nil
}

func (l *Logger) Info(message, ip, uid, traceID, method, path, module string, status int16, latency int64) {
	l.logger.Info(message, l.Field(ip, uid, traceID, method, path, module, status, latency)...)
}

func (l *Logger) Warn(message string, ip, uid, traceID, method, path, module string, status int16, latency int64) {
	l.logger.Warn(message, l.Field(ip, uid, traceID, method, path, module, status, latency)...)
}

func (l *Logger) Error(message string, ip, uid, traceID, method, path, module string, status int16, latency int64) {
	l.logger.Error(message, l.Field(ip, uid, traceID, method, path, module, status, latency)...)
}

func (l *Logger) Field(ip, uid, traceID, method, path, module string, status int16, latency int64) []zap.Field {
	return []zap.Field{
		zap.Int64("event_time", time.Now().UnixMilli()), // 建议用毫秒，CK 存储更精准
		zap.String("uid", uid),
		zap.String("ip", ip),
		zap.String("method", method),
		zap.String("path", path),
		zap.Int16("status", status),
		zap.Int64("latency", latency),
		zap.String("trace_id", traceID),
		zap.String("module", module),
	}
}
