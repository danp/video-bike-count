# video-bike-count

1. capture a stable video of bike counts, such as with an iPhone propped up in a window
1. install opencv, such as with [homebrew](https://github.com/hybridgroup/gocv#installation-1)
1. run this with `-video-file <that-file> -video-start-time 06:24:32`
1. find interesting frames in `out/`

## TODO

* make target rect configurable

## Other fun

Generate a supercut video from the files in `out/` with:

```
cd out && ffmpeg -framerate 30 -pattern_type glob -i '*.jpg' -c:v libx264 -pix_fmt yuv420p ../recap.mp4
```
