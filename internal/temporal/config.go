package temporal

type Config struct {
	Host          string `yaml:"host" validate:"required"`
	Port          int    `yaml:"port" validate:"required"`
	TaskQueueName string `yaml:"taskQueueName" validate:"required"`
}
