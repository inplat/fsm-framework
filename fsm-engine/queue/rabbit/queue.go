package rabbit

import (
	"github.com/streadway/amqp"
)

type Queue struct {
	q amqp.Queue
}

func (q *Queue) GetName() string {
	return q.q.Name
}
