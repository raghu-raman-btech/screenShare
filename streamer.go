package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"net"
	"time"

	"github.com/kbinani/screenshot"
)

// Settings
const (
	N_FRAMES          = 200
	FPS               = 10
	KEYFRAME_INTERVAL = FPS * 5
)

func rectBytes(x, y, w, h int) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, int32(x))
	_ = binary.Write(buf, binary.LittleEndian, int32(y))
	_ = binary.Write(buf, binary.LittleEndian, int32(w))
	_ = binary.Write(buf, binary.LittleEndian, int32(h))
	return buf.Bytes()
}

func colorEquals(c1, c2 color.Color) bool {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	const t = 30 * 257
	return absDiff(r1, r2) < t && absDiff(g1, g2) < t && absDiff(b1, b2) < t
}

func absDiff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

func compareImages(a, b image.Image) (int, int, int, int, bool) {
	bounds := a.Bounds()
	x0, y0 := bounds.Max.X, bounds.Max.Y
	x1, y1 := 0, 0
	found := false
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if !colorEquals(a.At(x, y), b.At(x, y)) {
				found = true
				if x < x0 {
					x0 = x
				}
				if y < y0 {
					y0 = y
				}
				if x > x1 {
					x1 = x
				}
				if y > y1 {
					y1 = y
				}
			}
		}
	}
	if found {
		return x0, y0, x1 - x0 + 1, y1 - y0 + 1, true
	}
	return 0, 0, 0, 0, false
}

func encodeBin(flag byte, x, y, w, h int, img image.Image) []byte {
	buf := new(bytes.Buffer)
	buf.Write([]byte{flag})
	buf.Write(rectBytes(x, y, w, h))
	region := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(region, region.Bounds(), img, image.Point{x, y}, draw.Src)
	_ = jpeg.Encode(buf, region, &jpeg.Options{Quality: 70})
	return buf.Bytes()
}

func main() {
	bounds := screenshot.GetDisplayBounds(0)
	WIDTH, HEIGHT := bounds.Dx(), bounds.Dy()
	fmt.Printf("Screen size: %dx%d\n", WIDTH, HEIGHT)

	conn, err := net.Dial("tcp", "localhost:8082")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	var prev image.Image
	for i := 0; i < N_FRAMES; i++ {
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			panic(err)
		}
		keyframe := i%KEYFRAME_INTERVAL == 0 || prev == nil

		var payload []byte
		if keyframe {
			payload = encodeBin('F', 0, 0, WIDTH, HEIGHT, img)
			fmt.Printf("Pushed FULL frame_%03d (%.1f KB)\n", i, float64(len(payload))/1024.0)
		} else {
			x, y, w, h, changed := compareImages(prev, img)
			if changed && w > 0 && h > 0 {
				payload = encodeBin('D', x, y, w, h, img)
				fmt.Printf("Pushed DIRTY frame_%03d (%.1f KB), rect: x=%d y=%d w=%d h=%d\n", i, float64(len(payload))/1024.0, x, y, w, h)
			} else {
				prev = img
				time.Sleep(time.Second / FPS)
				continue
			}
		}

		length := uint32(len(payload))
		binary.Write(conn, binary.BigEndian, length)
		conn.Write(payload)
		prev = img
		time.Sleep(time.Second / FPS)
	}
}
