module example.com/hello

go 1.13

replace example.com/sub => ./sub

require (
	example.com/signal v0.0.0-00010101000000-000000000000
	example.com/sub v0.0.0-00010101000000-000000000000
	github.com/pion/rtcp v1.2.6
	github.com/pion/webrtc/v3 v3.0.31
)

replace github.com/pion/webrtc/v3/examples/internal/signal => ./internal/signal

replace example.com/signal => ./internal/signal
