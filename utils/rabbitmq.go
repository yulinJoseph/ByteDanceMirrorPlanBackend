package utils

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"path"
	"runtime"
)

func ReadRabbitmqConfig() string {
	_, fileName, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("No caller information")
		return ""
	}
	fileName = path.Join(path.Dir(fileName), "../config/rabbitmq.json")

	viper.SetConfigFile(fileName)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Read config file error: %s\n", err)
		return ""
	}

	config := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		viper.GetString("user"),
		viper.GetString("passwd"),
		viper.GetString("host"),
		viper.GetString("port"))

	return config
}

type Rabbitmq struct {
	Conn       *amqp.Connection
	Channel    *amqp.Channel
	QueueName  string
	Exchange   string
	RoutingKey string
	MqUrl      string
}

func NewRabbitmq(queueName string, exchange string, routingKey string) *Rabbitmq {
	rabbitmq := Rabbitmq{
		QueueName:  queueName,
		Exchange:   exchange,
		RoutingKey: routingKey,
		MqUrl:      ReadRabbitmqConfig(),
	}
	var err error
	rabbitmq.Conn, err = amqp.Dial(rabbitmq.MqUrl)
	if err != nil {
		fmt.Printf("Dial rabbitmq error: %s\n", err)
		panic(err)
	}

	rabbitmq.Channel, err = rabbitmq.Conn.Channel()
	if err != nil {
		fmt.Printf("Channel rabbitmq error: %s\n", err)
		panic(err)
	}
	return &rabbitmq
}

var MQ4UserBuy *Rabbitmq
var MQ4SellTicket *Rabbitmq

func InitRabbitmq() {
	MQ4UserBuy = NewRabbitmq("user_buy_queue", "user_buy_exchange", "user_buy_key")
	MQ4SellTicket = NewRabbitmq("sell_ticket_queue", "sell_ticket_exchange", "sell_ticket_key")
	initRabbitmq(MQ4UserBuy)
	initRabbitmq(MQ4SellTicket)
}

func initRabbitmq(mq *Rabbitmq) {
	_, err := mq.Channel.QueueDeclare(
		mq.QueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Printf("QueueDeclare error: %s\n", err)
		return
	}

	err = mq.Channel.ExchangeDeclare(
		mq.Exchange,
		"direct",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Printf("ExchangeDeclare error: %s\n", err)
		return
	}

	err = mq.Channel.QueueBind(
		mq.QueueName,
		mq.RoutingKey,
		mq.Exchange,
		false,
		nil,
	)
	if err != nil {
		fmt.Printf("QueueBind error: %s\n", err)
		return
	}
}

func Publish(mq *Rabbitmq, exchange string, routingKey string, body string) error {
	err := mq.Channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})

	return err
}

func Consume(mq *Rabbitmq) (<-chan amqp.Delivery, error) {
	msg, err := mq.Channel.Consume(
		mq.QueueName,
		"",
		true,
		false,
		false,
		true,
		nil,
	)
	if err != nil {
		fmt.Printf("Consume error: %s\n", err)
		return nil, err
	}

	return msg, nil
}
