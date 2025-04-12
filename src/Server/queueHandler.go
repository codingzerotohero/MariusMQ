package main

import (
	"github.com/google/uuid"
)

type QueueHandler struct {
}

func (q *QueueHandler) GetQueueByName() {

}

func (q *QueueHandler) GetQueueById() {

}

func (q *QueueHandler) GetQueues() {

}

func (q *QueueHandler) DeleteQueue() {

}

func (q *QueueHandler) DeleteQueues() {

}

func (q *QueueHandler) CreateQueue() {

}

type MessageQueue struct {
	QueueName string
	Id        uuid.UUID
}
