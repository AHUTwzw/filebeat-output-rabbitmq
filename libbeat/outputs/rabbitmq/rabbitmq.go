package rabbitmqout

import (
	"context"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/outputs"
	"github.com/elastic/beats/v7/libbeat/outputs/codec"
	"github.com/elastic/beats/v7/libbeat/publisher"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	amqp "github.com/rabbitmq/amqp091-go"
	"net/url"
)

func init() {
	outputs.RegisterType("rabbitmq", rabbitmqout)
}

type rabbitmqOutput struct {
	log        *logp.Logger
	beat       beat.Info
	observer   outputs.Observer
	codec      codec.Codec
	conn       *amqp.Connection
	channel    *amqp.Channel
	exchange   string
	routingKey string
}

// rabbitmqout instantiates a new rabbitmq output instance.
func rabbitmqout(
	_ outputs.IndexManager,
	beat beat.Info,
	observer outputs.Observer,
	cfg *config.C,
) (outputs.Group, error) {
	config := defaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return outputs.Fail(err)
	}

	// disable bulk support in publisher pipeline
	cfg.SetInt("bulk_max_size", -1, -1)

	rbmqo := &rabbitmqOutput{
		log:      logp.NewLogger("rabbitmq"),
		beat:     beat,
		observer: observer,
	}
	if err := rbmqo.init(beat, config); err != nil {
		return outputs.Fail(err)
	}

	return outputs.Success(config.Queue, -1, 0, nil, rbmqo)
}

func (out *rabbitmqOutput) init(beat beat.Info, c RabbitmqConfig) error {
	connectionString := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", c.Username, url.QueryEscape(c.Password), c.Host, c.Port, c.Vhost)

	var err error

	out.codec, err = codec.CreateEncoder(beat, c.Codec)
	if err != nil {
		return err
	}

	out.conn, err = amqp.Dial(connectionString)
	if err != nil {
		out.log.Errorf("Failed to connect to RabbitMQ: %v", err)
		return err
	}

	out.channel, err = out.conn.Channel()
	if err != nil {
		out.log.Errorf("Failed to open a channel: %v", err)
		return err
	}

	err = out.channel.ExchangeDeclare(
		c.Exchange, // name
		"direct",   // type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		out.log.Errorf("Failed to declare an exchange: %v", err)
		return err
	}

	out.exchange = c.Exchange
	out.routingKey = c.Routing

	return nil
}

func (out *rabbitmqOutput) Close() error {
	return out.conn.Close()
}

func (out *rabbitmqOutput) Publish(_ context.Context, batch publisher.Batch) error {
	defer batch.ACK()

	st := out.observer
	events := batch.Events()
	st.NewBatch(len(events))

	dropped := 0
	for i := range events {
		event := &events[i]

		serializedEvent, err := out.codec.Encode(out.beat.Beat, &event.Content)
		if err != nil {
			if event.Guaranteed() {
				out.log.Errorf("Failed to serialize the event: %+v", err)
			} else {
				out.log.Warnf("Failed to serialize the event: %+v", err)
			}
			out.log.Debugf("Failed event: %v", event)

			dropped++
			continue
		}

		if err = out.send(serializedEvent); err != nil {
			st.WriteError(err)

			if event.Guaranteed() {
				out.log.Errorf("Sending event to RabbitMQ failed with: %+v", err)
			} else {
				out.log.Warnf("Sending event to RabbitMQ failed with: %+v", err)
			}

			dropped++
			continue
		}

		st.WriteBytes(len(serializedEvent) + 1)
	}

	st.ErrTooMany(dropped)
	st.AckedEvents(len(events) - dropped)

	return nil
}

func (out *rabbitmqOutput) String() string {
	return "RabbitMQ"
}

func (out *rabbitmqOutput) send(data []byte) error {
	err := out.channel.Publish(
		out.exchange,   //exchange
		out.routingKey, //routing key
		false,          //mandatory
		false,          //immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		})
	if err != nil {
		out.log.Errorf("Failed to publish a message: %v", err)
		return err
	}
	out.log.Debugf("Send a msg: %s", string(data))
	return nil
}
