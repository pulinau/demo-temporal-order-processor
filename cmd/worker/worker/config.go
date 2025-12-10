package worker

type Config struct {
	Temporal TemporalConfig
}

type TemporalConfig struct {
	Host          string
	Port          int
	TaskQueueName string
}
