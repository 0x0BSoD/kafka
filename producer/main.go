package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/0x0bsod/goLogz"
	"github.com/0x0bsod/kafka/common"
	"github.com/segmentio/kafka-go"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var confPath string

func init() {
	flag.StringVar(&confPath, "config", "./producer.yml", "path to config in yml format")
	flag.StringVar(&confPath, "c", "./producer.yml", "path to config in yml format")
}

func main() {
	logs, err := goLogz.Init([]goLogz.ParameterItem{
		{
			Level:     "ACTION",
			OutHandle: "STDOUT",
			LineNum:   false,
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	logs.Colors = true

	flag.Parse()
	if confPath == "" {
		logs.Error("Please provide yaml config file by using -c,-config option")
		return
	}

	var config common.Config
	var ctx common.CustomContext

	err = config.ParseConfig(confPath)
	if err != nil {
		logs.Error(err.Error())
		return
	}

	ctx.BackgroundCtx = context.Background()
	ctx.Logger = logs
	ctx.Config = &config

	ctx.Logger.Info("Starting producer")
	produce(ctx)
}

type keyCounter struct {
	mu         sync.Mutex
	currentKey int
}

func (kc *keyCounter) Get() int {
	kc.mu.Lock()
	k := kc.currentKey
	kc.currentKey++
	kc.mu.Unlock()
	return k
}

func produce(ctx common.CustomContext) {
	// initialize a counter for keys
	var key keyCounter
	key.currentKey = 1
	var wg sync.WaitGroup

	for i := 0; i <= runtime.NumCPU(); i++ {
		wg.Add(1)
		go func(ID int, kc *keyCounter) {
			w := kafka.NewWriter(kafka.WriterConfig{
				Brokers: ctx.Config.Brokers,
				Topic:   ctx.Config.TopicI,
			})
			ctx.Logger.Info(fmt.Sprintf("[!] producer %d connected to %s topic", ID, ctx.Config.TopicI))

			defer wg.Done()
			for {
				key := kc.Get()
				err := w.WriteMessages(ctx.BackgroundCtx, kafka.Message{
					Key: []byte(strconv.Itoa(key)),
					// create an arbitrary message payload for the value
					Value: []byte(strconv.FormatInt(time.Now().Unix(), 10)),
				})
				if err != nil {
					ctx.Logger.Error("could not write message " + err.Error())
					return
				}
				// log a confirmation once the message is written
				ctx.Logger.Custom("ACTION", fmt.Sprintf("--> [%d] writes: %d", ID, key))
				time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
			}
		}(i, &key)
	}
	wg.Wait()
}
