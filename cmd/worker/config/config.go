package config

type WorkerConfig struct {
	Temporal TemporalConfig
}

type TemporalConfig struct {
	Host          string
	Port          int
	TaskQueueName string
}
