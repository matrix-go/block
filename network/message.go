package network

type GetStatusMessage struct {
}

func NewGetStatusMessage() *GetStatusMessage {
	return &GetStatusMessage{}
}

type StatusMessage struct {
	ID      string // id of server
	Version uint32
	Height  uint64
}

func NewStatusMessage() *StatusMessage {
	return &StatusMessage{}
}

type GetBlocksMessage struct {
	From uint64 // height
	To   uint64 // height, if to is 0, the max height will return
}

func NewGetBlocksMessage(from, to uint64) *GetBlocksMessage {
	return &GetBlocksMessage{
		From: from,
		To:   to,
	}
}
