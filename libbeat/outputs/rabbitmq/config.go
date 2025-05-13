package rabbitmqout

import (
	"github.com/elastic/beats/v7/libbeat/outputs/codec"
	"github.com/elastic/elastic-agent-libs/config"
)

type RabbitmqConfig struct {
	Codec    codec.Config     `RabbitmqConfig:"codec"`
	Host     string           `RabbitmqConfig:"host"`
	Port     uint             `RabbitmqConfig:"port"`
	Username string           `RabbitmqConfig:"username"`
	Password string           `RabbitmqConfig:"password"`
	Vhost    string           `RabbitmqConfig:"vhost"`
	Exchange string           `RabbitmqConfig:"exchange"`
	Routing  string           `RabbitmqConfig:"routing"`
	Queue    config.Namespace `config:"queue"`
}

var (
	defaultConfig = RabbitmqConfig{
		Host:     "localhost",
		Port:     5672,
		Username: "guest",
		Password: "guest",
		Vhost:    "",
		Exchange: "filebeat-exc",
		Routing:  "filebeat",
	}
)

func (c *RabbitmqConfig) Validate() error {
	return nil
}
