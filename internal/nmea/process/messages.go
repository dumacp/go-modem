package process

type MsgData struct {
	Data string
}
type MsgTick struct {
}
type MsgSendFrame struct {
	Topic string
	Data  *processedData
}
