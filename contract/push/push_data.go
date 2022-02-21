package push

import (
	"github.com/no-src/gofs/action"
	"github.com/no-src/gofs/contract"
)

// PushData the request data of the push api
type PushData struct {
	contract.FileInfo
	// Action the action of file change
	Action action.Action `json:"action"`
}