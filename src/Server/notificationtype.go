package main

type NotificationType string

const (
	CREATE_QUEUE_FAIL_ALREADY_EXISTS          NotificationType = "CREATE_QUEUE_FAIL_ALREADY_EXISTS"
	SUBSCRIBE_QUEUE_FAIL_DOES_NOT_EXIST       NotificationType = "SUBSCRIBE_QUEUE_FAIL_DOES_NOT_EXIST"
	PUBLISH_MESSAGE_FAIL_QUEUE_DOES_NOT_EXIST NotificationType = "PUBLISH_MESSAGE_FAIL_QUEUE_DOES_NOT_EXIST"
	DELETE_QUEUE_FAIL_DOES_NOT_EXIST          NotificationType = "DELETE_QUEUE_FAIL_DOES_NOT_EXIST"
	QUEUE_DELETED                             NotificationType = "QUEUE_DELETED"
	SEND_MESSAGE                              NotificationType = "SEND_MESSAGE"
)
