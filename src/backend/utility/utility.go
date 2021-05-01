package utility

import (
	"encoding/json"
	"syscall/js"
)

func ProgressUpdate(progress float32, event string, workerId int) {
	data := struct {
		Progress float32 `json:"progress"`
		Event    string  `json:"event"`
		WorkerID int     `json:"workerId"`
	}{
		progress,
		event,
		workerId,
	}

	raw, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	js.Global().Call("progressUpdate", string(raw))
}
