package main

import (
	"flag"
	"go-mysql-kafka/conf"
	"go-mysql-kafka/gkafka"
	"go-mysql-kafka/mapper"
	"go-mysql-kafka/sync_manager"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var cfg = flag.String("cfg", "app.toml", "setting up the configuration file")

func main() {
	var err error
	flag.Parse()

	conf.Setup(*cfg)
	// gredis.Setup()

	c := conf.Config

	// 创建一个信号chan
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		os.Kill,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGILL,
		syscall.SIGTRAP,
		syscall.SIGABRT,
		syscall.SIGBUS,
		syscall.SIGFPE,
		syscall.SIGKILL,
		syscall.SIGSEGV,
		syscall.SIGPIPE,
		syscall.SIGALRM,
		syscall.SIGTERM)

	// 初始化存储binlog位置
	positionHolder := sync_manager.NewFilePositionHolder(c.SourceDB.DataDir)

	kafkaProducer, err := gkafka.NewKafka(c)
	if err != nil {
		log.Fatalf("init kafka producer err: %+v", err)
	}

	//kafkaProducer.SendMessageTest()

	// 初始化分表分库的配置
	rowsMapper := mapper.NewDRDSMapper(c)

	var sm *sync_manager.SyncManager

	sm, err = sync_manager.NewSyncManager(c, positionHolder, rowsMapper, kafkaProducer)
	if err != nil {
		log.Fatalf("init sync manager err: %+v", err)
	}

	done := make(chan struct{}, 1)

	go func() {
		err = sm.Run()
		if err != nil {
			log.Fatalf("sync manager run err: %v", err)
		}

		done <- struct{}{}
		log.Infof("run end")

	}()

	select {
	case n := <-sc:
		log.Infof("receive signal %v, closing", n)
		//TODO 临时写一下，之后应该把多个manager等context归总到一个进行监听
	case <-sm.Ctx.Done():
		log.Infof("context is done with %v, closing", sm.Ctx.Err())
	}

	sm.Close()
	kafkaProducer.Close()
	<-done
	log.Infof("sync manager is stop")
}
