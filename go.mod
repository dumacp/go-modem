module github.com/dumacp/go-modem

go 1.20

replace github.com/dumacp/omvz => ../omvz

//replace github.com/dumacp/go-logs => ../go-logs

require (
	github.com/AsynkronIT/protoactor-go v0.0.0-20210520041424-43065ace108f
	github.com/dumacp/go-logs v0.0.2-0.20241119230451-ac5d49be15ce
	github.com/dumacp/gpsnmea v0.0.0-20201110195359-2994f05cfb52
	github.com/dumacp/omvz v0.0.0-00010101000000-000000000000
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/gogo/protobuf v1.3.2
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551
	github.com/looplab/fsm v0.2.0
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
)

require (
	github.com/Workiva/go-datastructures v1.0.53 // indirect
	github.com/brian-armstrong/gpio v0.0.0-20181227042754-72b0058bbbcb // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/orcaman/concurrent-map v0.0.0-20190107190726-7ed82d9cb717 // indirect
	github.com/yanatan16/itertools v0.0.0-20160513161737-afd1891e0c4f // indirect
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb // indirect
	golang.org/x/sys v0.0.0-20210511113859-b0526f3d8744 // indirect
)
