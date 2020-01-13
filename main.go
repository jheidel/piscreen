package main

import (
	"piscreen/lcd"
	"piscreen/runner"
	"os"
	"sync"
	"context"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Start piscreen")

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	display, err := lcd.NewDisplay(ctx, wg)
	if err != nil {
		panic(err)
	}

	runner := &runner.Runner{
		Display: display,
		Path: "/home/pi/display.sh",
	}
	runner.Run(ctx, wg)

	log.Warnf("piscreen running. Waiting for signal.")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	log.Warnf("%v, exiting...", sig)
	cancel()
	wg.Wait()
	log.Infof("Exited.")
}
