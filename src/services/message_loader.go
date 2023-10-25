package services

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"go_test/pqueue"
	"go_test/services/dto"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type MessageLoader struct {
	wsConnection *websocket.Conn
	done         chan struct{}
	queue        *pqueue.PriorityQueueMutex
	sender       *MessageSender
	queuedCount  int
	instantCount int
}

func NewMessageLoader(wsConnection *websocket.Conn, done chan struct{}, queue *pqueue.PriorityQueueMutex, sender *MessageSender) *MessageLoader {
	return &MessageLoader{
		wsConnection: wsConnection,
		done:         done,
		queue:        queue,
		sender:       sender,
	}
}

func (ml *MessageLoader) Load() {
	go func() {
		defer close(ml.done)

		// var prevId int
		// var isPrevIdSet bool

		for {
			_, message, err := ml.wsConnection.ReadMessage()
			if err != nil {
				fmt.Println("readMessage error:", err)
				return
			}

			var msg dto.Message
			err = json.Unmarshal(message, &msg)
			if err != nil {
				log.Println("unmarshal error", err)
				return
			}

			item := &pqueue.Item{Value: msg.Text, Priority: msg.Id, Created: time.Now()}

			// Here is trying of optimization if we have consequenced msg.Id
			// The idea is in sending message instantly without queueing it.
			// But it have many courner cases when we trying handle messages in different goroutines asyncroniously.

			// shouldSendInstantly := false
			// if !isPrevIdSet {
			// 	shouldSendInstantly = false
			// }
			// if prevId+1 == msg.Id {
			// 	shouldSendInstantly = true
			// }

			// if shouldSendInstantly {
			// 	prevId = msg.Id
			// 	err := ml.sender.SendOneMessage(item)
			// 	if err != nil {
			// 		log.Println("SendOneMessage error", err)
			// 		return
			// 	}
			//  ml.instantCount++
			// 	continue
			// }
			// ml.queuedCount++

			heap.Push(ml.queue, item)
		}
	}()
}

func (ml *MessageLoader) GetQueuedCount() int {
	return ml.queuedCount
}

func (ml *MessageLoader) GetInstantCount() int {
	return ml.instantCount
}
