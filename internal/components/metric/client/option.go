package metric_sdk

type CounterVectorOptions struct {
	Name string 
	Help string 
	Labels []string
}

type HistogramVectorOptions  struct {
	Name string
	Help string 
	Labels []string
	BucketsBoundaries []float64
}