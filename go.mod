module github.com/dumacp/go-modem

go 1.16

replace github.com/dumacp/omvz => ../omvz

replace github.com/dumacp/go-logs => ../go-logs

require (
	github.com/AsynkronIT/protoactor-go v0.0.0-20210520041424-43065ace108f
	github.com/dumacp/go-logs v0.0.0-20211122205852-dfcc5685457f
	github.com/dumacp/gpsnmea v0.0.0-20201110195359-2994f05cfb52
	github.com/dumacp/omvz v0.0.0-00010101000000-000000000000
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/gogo/protobuf v1.3.2
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551
	github.com/looplab/fsm v0.2.0
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
)
