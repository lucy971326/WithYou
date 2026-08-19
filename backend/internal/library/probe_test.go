package library

import "testing"

func TestBrowserSafe(t *testing.T) {
	cases := []struct {
		stream videoStream
		want   bool
	}{
		{videoStream{CodecName: "h264", PixFmt: "yuv420p"}, true},
		{videoStream{CodecName: "h264", PixFmt: "yuvj420p"}, true},
		{videoStream{CodecName: "h264", PixFmt: "yuv420p10le"}, false},
		{videoStream{CodecName: "hevc", PixFmt: "yuv420p"}, false},
		{videoStream{CodecName: "mpeg4", PixFmt: "yuv420p"}, false},
	}
	for _, tc := range cases {
		if got := browserSafe(tc.stream); got != tc.want {
			t.Fatalf("browserSafe(%+v)=%v want %v", tc.stream, got, tc.want)
		}
	}
}
