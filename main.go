package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"time"

	"gocv.io/x/gocv"
)

var (
	minimumArea    = flag.Int("minimum-contour-area", 10000, "minimum contour area to consider")
	videoFile      = flag.String("video-file", "", "video file to read")
	videoStartTime = flag.String("video-start-time", "", "real start time of video, in HH:MM:SS")
	outDir         = flag.String("out-dir", "out", "directory to write interesting frames")
)

const MinimumArea = 10000

func main() {
	flag.Parse()
	if *videoFile == "" {
		fmt.Fprintln(os.Stderr, "need -video-file")
		os.Exit(1)
	}

	var videoStart time.Time
	if *videoStartTime != "" {
		t, err := time.Parse("15:04:05", *videoStartTime)
		if err != nil {
			fmt.Fprintln(os.Stderr, "bad -video-start-time:", err)
			os.Exit(1)
		}
		videoStart = t
	}

	if err := os.Mkdir(*outDir, 0755); err != nil && !os.IsExist(err) {
		fmt.Fprintf(os.Stderr, "error creating -out-dir %q: %s\n", *outDir, err)
		os.Exit(1)
	}

	video, err := gocv.VideoCaptureFile(*videoFile)
	if err != nil {
		fmt.Printf("Error opening video capture device: %v\n", *videoFile)
		return
	}
	defer video.Close()

	window := gocv.NewWindow("Motion Window")
	defer window.Close()

	img := gocv.NewMat()
	defer img.Close()

	imgDelta := gocv.NewMat()
	defer imgDelta.Close()

	imgThresh := gocv.NewMat()
	defer imgThresh.Close()

	mog2 := gocv.NewBackgroundSubtractorMOG2()
	defer mog2.Close()

	target := image.Rect(660, 0, 680, 720)

	for {
		if ok := video.Read(&img); !ok {
			fmt.Printf("Error cannot read device %d\n", os.Args[1])
			return
		}
		if img.Empty() {
			continue
		}

		posSec := int(video.Get(gocv.VideoCapturePosMsec) / 1000)
		frame := int(video.Get(gocv.VideoCapturePosFrames))

		// first phase of cleaning up image, obtain foreground only
		mog2.Apply(img, &imgDelta)

		// remaining cleanup of the image to use for finding contours.
		// first use threshold
		gocv.Threshold(imgDelta, &imgThresh, 25, 255, gocv.ThresholdBinary)

		// then dilate
		kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
		defer kernel.Close()
		gocv.Dilate(imgThresh, &imgThresh, kernel)

		// now find contours
		contours := gocv.FindContours(imgThresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)
		var interesting bool
		for _, c := range contours {
			area := gocv.ContourArea(c)
			if area < float64(*minimumArea) {
				continue
			}

			rect := gocv.BoundingRect(c)
			if !rect.Overlaps(target) {
				continue
			}

			interesting = true
			break
		}

		if interesting {
			fn := fmt.Sprintf(*outDir+"/frame-%05ds-%05df.jpg", posSec, frame)

			if !videoStart.IsZero() {
				t := videoStart.Add(time.Duration(posSec) * time.Second)
				gocv.PutText(&img, fmt.Sprintf("%s %ds f%d", t.Format("15:04:05"), posSec, frame), image.Pt(10, 40), gocv.FontHersheyPlain, 1.5, color.RGBA{255, 255, 255, 0}, 2)
			}

			fmt.Println(fn)
			if !gocv.IMWrite(fn, img) {
				panic("couldn't write " + fn)
			}
		}
	}
}
