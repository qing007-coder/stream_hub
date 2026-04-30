package infra

type MetricsManager struct {

}

type Metrics struct {
	ServiceName string 
	Label []string
	Count uint64
}