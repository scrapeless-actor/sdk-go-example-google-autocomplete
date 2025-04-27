package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/scrapeless-ai/scrapeless-actor-sdk-go/scrapeless"
	proxyModel "github.com/scrapeless-ai/scrapeless-actor-sdk-go/scrapeless/proxy"
)

var (
	client *http.Client
)

func main() {
	actor := scrapeless.New(scrapeless.WithProxy(), scrapeless.WithStorage())
	defer actor.Close()
	// Get proxy
	proxy, err := actor.Proxy.Proxy(context.TODO(), proxyModel.ProxyActor{
		Country:         "us",
		SessionDuration: 10,
	})
	if err != nil {
		panic(err)
	}
	parse, err := url.Parse(proxy)
	if err != nil {
		panic(err)
	}
	// Set up proxy using Golang's native HTTP
	client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(parse)}}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-quit:
			log.Info("quit success")
			return
		default:
			doCrawl(actor, client)
		}
	}
}

func doCrawl(actor *scrapeless.Actor, client *http.Client) {
	queueResp, err := actor.Storage.GetQueue().Pull(context.TODO(), 1)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	if len(queueResp) == 0 {
		log.Info("no task")
		time.Sleep(time.Second)
		return
	}
	req, err := http.NewRequest(http.MethodGet, queueResp[0].Payload, nil)
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("sec-ch-ua-full-version-list", `"Chromium";v="134.0.6998.166", "Not:A-Brand";v="24.0.0.0", "Google Chrome";v="134.0.6998.166""`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-ch-ua-platform-version", "10.0.0")
	req.Header.Set("sec-ch-ua-wow64", "?0")
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-site", "none")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("upgrade-insecure-requests", "1")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36")
	req.Header.Set("x-browser-channel", "stable")
	req.Header.Set("x-browser-year", "2025")
	if err != nil {
		log.Error(err)
		panic(err)
	}
	do, err := client.Do(req)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	body, _ := io.ReadAll(do.Body)
	objectId, err := actor.Storage.GetObject().Put(context.TODO(), queueResp[0].Payload+".json", body)
	if err != nil {
		log.Error(err)
		panic(err)
	}
	log.Info("success, objectId:", objectId)
	get, _ := actor.Storage.GetObject().Get(context.TODO(), objectId)
	log.Info(string(get))
	if err := actor.Storage.GetQueue().Ack(context.TODO(), queueResp[0].ID); err != nil {
		log.Error(err)
	}
}
