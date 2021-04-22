
start:
	GOOS=js GOARCH=wasm go build -o ./build/main.wasm .
  
restart: start
	@echo "REBUILD" && printf '%*s\n' "40" '' | tr ' ' -
  
# Serve task will run fswatch monitor and performs restart task if any source file changed. Before serving it will execute start task.
serve: start
	fswatch -or --event=Updated /workspace | \
	xargs -n1 -I {} make restart
  
# .PHONY is used for reserving tasks words
.PHONY: start before stop restart serve
