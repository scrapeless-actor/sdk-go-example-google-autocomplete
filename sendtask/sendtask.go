package main

import (
	"context"
	"github.com/scrapeless-ai/scrapeless-actor-sdk-go/scrapeless"
	"github.com/scrapeless-ai/scrapeless-actor-sdk-go/scrapeless/storage/queue"
	log "github.com/sirupsen/logrus"
)

func main() {
	sl := scrapeless.New(scrapeless.WithStorage())
	defer sl.Close()

	push, _ := sl.Storage.GetQueue().Push(context.TODO(), queue.PushQueue{
		Name:     "task",
		Payload:  []byte("https://www.google.com/complete/search?q=Coffee&client=chrome&dpr=1"),
		Retry:    0,
		Timeout:  0,
		Deadline: 0,
	})
	log.Info(push)
}
