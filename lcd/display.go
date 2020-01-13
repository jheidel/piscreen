package lcd

import (
	"sync"
	"context"
	"strings"
	"fmt"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"

	"github.com/goiot/devices/monochromeoled"
	"golang.org/x/exp/io/i2c"

	log "github.com/sirupsen/logrus"
)

type Display struct {
	oled *monochromeoled.OLED

	textc chan string
}

func NewDisplay(ctx context.Context, wg *sync.WaitGroup) (*Display, error) {
	oled, err := monochromeoled.Open(&i2c.Devfs{Dev: "/dev/i2c-1"})
	if err != nil {
		return nil, fmt.Errorf("opening oled: %v", err)
	}
	if err := oled.On(); err != nil {
		return nil, fmt.Errorf("turning on oled: %v", err)
	}
	if err := oled.Clear(); err != nil {
		return nil, fmt.Errorf("clearing oled: %v", err)
	}

	wg.Add(1)
	d := &Display{
		oled: oled,
		textc: make(chan string),
	}

	go func() {
		defer wg.Done()
		d.loop(ctx)
		if err := d.cleanup(); err != nil {
			log.Errorf("Display cleanup failed: %v", err)
		}
	}()
	return d, nil
}

func (d *Display) cleanup() error {
	defer d.oled.Close()
	if err := d.oled.Clear(); err != nil {
		return err
	}
	if err := d.oled.Off(); err != nil {
		return err
	}
	return nil
}

func addLabel(img *image.RGBA, label string, x, y int) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.RGBA{255, 255, 255, 255}),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(label)
}

//  func addCenteredLabel(img *image.RGBA, label string, y int) {
//  	d := &font.Drawer{
//  		Dst:  img,
//  		Src:  image.NewUniform(color.RGBA{255, 255, 255, 255}),
//  		Face: basicfont.Face7x13,
//  		Dot:  fixed.P(0, y),
//  	}
//  	advance := d.MeasureString(label)
//  	d.Dot.X = fixed.I(128)/2 - advance/2
//  	d.DrawString(label)
//  }

func (d *Display) DrawText(text string) {
	d.textc <- text
}

func (d *Display) draw(text string) error {
	img := image.NewRGBA(image.Rect(0, 0, 128, 64))

	for i, line := range strings.Split(strings.TrimSpace(text), "\n") {
		line = strings.Replace(line, "\t", " ", -1)
		addLabel(img, line, 0, 13 * (i+1))
	}

	if err := d.oled.SetImage(0, 0, img); err != nil {
		return err
	}
	if err := d.oled.Draw(); err != nil {
		return err
	}
	return nil
}

func (d *Display) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case text := <-d.textc:
			if err := d.draw(text); err != nil {
				log.Warnf("Error drawing display: %v", err)
			}
		}
	}
}
