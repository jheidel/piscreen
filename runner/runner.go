package runner

import(
	"sync"
	"context"
	"os/exec"
	"time"
	"piscreen/lcd"

	log "github.com/sirupsen/logrus"
)

var Interval = time.Second

type Runner struct {
	Display *lcd.Display
	Path string
}

func (r *Runner) runProcess() {
	c := exec.Command(r.Path)
	o, err := c.CombinedOutput()
	if err != nil {
		log.Errorf("Subprocess failed: %v", err)
		return
	}

	r.Display.DrawText(string(o))
}

func (r *Runner) Run(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		r.runProcess()
		t := time.NewTicker(Interval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				r.runProcess()
			}
		}
	}()
}
