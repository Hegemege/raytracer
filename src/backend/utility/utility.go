package utility

import (
	"encoding/json"
	"syscall/js"
)

func ProgressUpdate(progress float32, event string, taskId int) {
	data := struct {
		Progress float32 `json:"progress"`
		Event    string  `json:"event"`
		TaskID   int     `json:"taskId"`
	}{
		progress,
		event,
		taskId,
	}

	raw, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	js.Global().Call("progressUpdate", string(raw))
}
