
start:
	date +%Y-%m-%dT%H:%M:%S
	GOOS=js GOARCH=wasm go build -o ./build/main.wasm .
	md5sum ./build/main.wasm
  
restart: start
	@echo "REBUILD" && printf '%*s\n' "40" '' | tr ' ' -
  
# Serve task will run fswatch monitor and performs restart task if any source file changed. Before serving it will execute start task.
serve: start
	fswatch -x --event Updated -m poll_monitor /workspace/*.go /workspace/*/*.go | \
	xargs -n1 -I {} make restart
  
# .PHONY is used for reserving tasks words
.PHONY: start restart serve
