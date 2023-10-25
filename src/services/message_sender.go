package services

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"go_test/pqueue"
	"go_test/services/dto"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const QUEUE_WAITING_LEN = 300
const QUEUE_WAITING_TIME = 10 * time.Millisecond
const MESSAGE_SEND_PAUSE_AFTER = 100 * time.Microsecond // Required to give a time for working MessageLoader and heapify queue.
const MESSAGES_SENT_TO_SUCCESS = 5000

type MessageSender struct {
	wsConnection *websocket.Conn
	done         chan struct{}
	queue        *pqueue.PriorityQueueMutex
	maxId        int
	sentCount    int
	minLatency   time.Duration
}

func NewMessageSender(wsConnection *websocket.Conn, done chan struct{}, queue *pqueue.PriorityQueueMutex) *MessageSender {
	return &MessageSender{
		wsConnection: wsConnection,
		done:         done,
		queue:        queue,
		minLatency:   1<<63 - 1,
	}
}

func (ml *MessageSender) Send() {
	go func() {
		defer close(ml.done)

		for {
			if ml.queue.Len() < QUEUE_WAITING_LEN {
				log.Println("Sending sleep... Zzz")
				time.Sleep(QUEUE_WAITING_TIME)
				continue
			}

			item := heap.Pop(ml.queue).(*pqueue.Item)

			err := ml.SendOneMessage(item)
			if err != nil {
				log.Println("SendOneMessage error: ", err)
				return
			}
			time.Sleep(MESSAGE_SEND_PAUSE_AFTER)
		}
	}()
}

func (ml *MessageSender) SendOneMessage(item *pqueue.Item) error {
	b, err := json.Marshal(dto.Message{Id: item.Priority, Text: item.Value})

	if ml.maxId < item.Priority {
		ml.maxId = item.Priority
	} else {
		return fmt.Errorf("Unordered message: %d, %d", ml.maxId, item.Priority)
	}

	if err != nil {
		return fmt.Errorf("jsonMarshal error: %w", err)
	}

	output := []byte(strings.ToLower(string(b)))

	err = ml.wsConnection.WriteMessage(websocket.TextMessage, output)
	if err != nil {
		return fmt.Errorf("WriteMessage error: %w", err)
	}

	diffTime := time.Now().Sub(item.Created)
	if ml.minLatency > diffTime {
		ml.minLatency = diffTime
	}

	ml.sentCount++

	log.Printf("sent: %s", b)

	if ml.sentCount > MESSAGES_SENT_TO_SUCCESS {
		return fmt.Errorf("Successfully sent %d messages", MESSAGES_SENT_TO_SUCCESS)
	}

	return nil
}

func (ml *MessageSender) GetMaxId() int {
	return ml.maxId
}

func (ml *MessageSender) GetSentCount() int {
	return ml.sentCount
}

func (ml *MessageSender) GetMinLatency() time.Duration {
	return ml.minLatency
}
