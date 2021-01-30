package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/0x0bsod/goLogz"
	"github.com/0x0bsod/kafka/common"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var (
	confPath string
)

func init() {
	flag.StringVar(&confPath, "config", "./consumer.yml", "path to config in yml format")
	flag.StringVar(&confPath, "c", "./consumer.yml", "path to config in yml format")
}

type workersStruct map[int]chan bool

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
	stackIn := common.NewStack()
	stackOut := common.NewStack()

	err = config.ParseConfig(confPath)
	if err != nil {
		logs.Error(err.Error())
		return
	}

	workersIdM := 1
	workersIdP := 1
	spawnEdgeM := 20
	spawnEdgeP := 20

	ctx.BackgroundCtx = context.Background()
	ctx.Logger = logs
	ctx.StackIn = stackIn
	ctx.StackOut = stackOut
	ctx.Config = &config

	workersM := make(workersStruct)
	workersP := make(workersStruct)

	var stopCh chan bool
	ctx.Logger.Info("Starting consumer")
	go consume(ctx)
	ctx.Logger.Info("Starting mutator 1")
	workersM[workersIdM] = stopCh
	go mutate(ctx, workersIdM, workersM[workersIdM])
	// prometheus `MutatorsCount`
	MutatorsCount.Inc()

	ctx.Logger.Info("Starting producer 1")
	workersP[workersIdP] = stopCh
	go produce(ctx, workersIdP, workersP[workersIdP])
	// prometheus `ProducersCount`
	ProducersCount.Inc()

	// for prometheus
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		panic(http.ListenAndServe(fmt.Sprintf(":%d", ctx.Config.PrometheusPort), nil))
	}()

	for {
		if ctx.StackIn.Len() > spawnEdgeM {
			ctx.Logger.Warning(fmt.Sprintf("Stack contains more than %d elements, starting new mutate worker", spawnEdgeM))
			workersIdM++
			spawnEdgeM = spawnEdgeM * 2
			var stopCh chan bool
			workersM[workersIdM] = stopCh
			go mutate(ctx, workersIdM, workersM[workersIdM])
			// prometheus `MutatorsCount`
			MutatorsCount.Inc()
			ctx.Logger.Warning(fmt.Sprint("Num of mutate workers: ", workersIdM))
		}
		if ctx.StackIn.Len() < spawnEdgeM {
			if workersIdM != 1 {
				ctx.Logger.Warning(fmt.Sprintf("Stack contains less than %d elements, stopping mutate worker", spawnEdgeM))
				workersM[workersIdM] <- true
				workersIdM--
				spawnEdgeM = spawnEdgeM / 2
				delete(workersM, workersIdM)
				// prometheus `MutatorsCount`
				MutatorsCount.Dec()
				ctx.Logger.Warning(fmt.Sprint("Num of mutate workers: ", workersIdM))
			}
		}

		if ctx.StackOut.Len() > spawnEdgeP {
			ctx.Logger.Warning(fmt.Sprintf("Stack contains more than %d elements, starting new producer worker", spawnEdgeP))
			workersIdP++
			spawnEdgeP = spawnEdgeP * 2
			var stopCh chan bool
			workersP[workersIdP] = stopCh
			go produce(ctx, workersIdP, workersP[workersIdP])
			// prometheus `ProducersCount`
			ProducersCount.Inc()
			ctx.Logger.Warning(fmt.Sprint("Num of producer workers: ", workersIdP))
		}
		if ctx.StackOut.Len() < spawnEdgeP {
			if workersIdP != 1 {
				ctx.Logger.Warning(fmt.Sprintf("Stack contains less than %d elements, sopping producer worker", spawnEdgeM))
				workersP[workersIdP] <- true
				workersIdP--
				spawnEdgeP = spawnEdgeP / 2
				delete(workersP, workersIdP)
				// prometheus `ProducersCount`
				ProducersCount.Dec()
				ctx.Logger.Warning(fmt.Sprint("Num of producer workers: ", workersIdP))
			}
		}
	}

}

func consume(ctx common.CustomContext) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  ctx.Config.Brokers,
		Topic:    ctx.Config.TopicI,
		GroupID:  fmt.Sprintf("mutators"),
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	ctx.Logger.Info(fmt.Sprintf("[!] consumer connected to %s topic", ctx.Config.TopicI))
	var wg sync.WaitGroup

	for i := 1; i <= runtime.NumCPU(); i++ {
		go func(ID int) {
			// prometheus `ConsumersCount`
			ConsumersCount.Inc()
			wg.Add(1)
			defer func() {
				ConsumersCount.Dec()
				wg.Done()
			}()
			for {
				msg, err := r.ReadMessage(ctx.BackgroundCtx)
				if err != nil {
					ctx.Logger.Error("could not fetch message " + err.Error())
				}
				ctx.Logger.Custom("ACTION", fmt.Sprintf("<-- [%d] received: %s key %s", ID, string(msg.Value), string(msg.Key)))

				ctx.StackIn.Push(&common.Node{
					Key:   msg.Key,
					Value: msg.Value,
				})
				// prometheus `InputStackSize`
				InputStackSize.Inc()
			}
		}(i)
	}
	wg.Wait()
}

func mutate(ctx common.CustomContext, ID int, stopCh chan bool) {
	for {
		select {
		case _ = <-stopCh:
			ctx.Logger.Info(fmt.Sprintf("[!] mutator %d stopped", ID))
		default:
			msg := ctx.StackIn.Get()
			if msg != nil {
				// prometheus `InputStackSize`
				InputStackSize.Dec()
				tUnix, err := strconv.ParseInt(string(msg.Value), 10, 64)
				if err != nil {
					ctx.Logger.Error(err.Error())
					return
				}

				timeT := time.Unix(tUnix, 0).Format(time.RFC3339)

				ctx.Logger.Info(fmt.Sprintf("-- [%d] mutated: %s", ID, timeT))

				ctx.StackOut.Push(&common.Node{
					Key:   msg.Key,
					Value: []byte(timeT),
				})
				// prometheus `OutputStackSize`
				OutputStackSize.Inc()
			}
		}
	}
}

func produce(ctx common.CustomContext, ID int, stopCh chan bool) {

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: ctx.Config.Brokers,
		Topic:   ctx.Config.TopicO,
	})

	ctx.Logger.Info(fmt.Sprintf("[!] producer %d connected to %s topic", ID, ctx.Config.TopicO))

	for {
		select {
		case _ = <-stopCh:
			ctx.Logger.Info(fmt.Sprintf("[!] producer %d stopped", ID))
		default:
			msg := ctx.StackOut.Get()
			if msg != nil {
				// prometheus `OutputStackSize`
				OutputStackSize.Dec()
				err := w.WriteMessages(ctx.BackgroundCtx, kafka.Message{
					Key:   msg.Key,
					Value: msg.Value,
				})
				if err != nil {
					ctx.Logger.Error(err.Error())
					return
				}
				ctx.Logger.Custom("ACTION", fmt.Sprintf("--> [%d] writes: %s key %s", ID, string(msg.Value), string(msg.Key)))
			}
		}

	}
}
