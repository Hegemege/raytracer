// From https://qvault.io/javascript/running-go-in-the-browser-with-wasm-and-web-workers/

// eslint-disable-next-line no-undef
importScripts("wasm_exec.js");

// eslint-disable-next-line no-unused-vars
function progressUpdate(params) {
  postMessage({
    progressUpdate: true,
    data: JSON.parse(params),
  });
}

{
  let go = null;
  let workerId = null;
  let renderFunc = null;

  onmessage = async (e) => {
    // Sending logging events to main thread
    let log = (wid, ...args) => {
      postMessage({
        logMessage: true,
        message: args.join(" "),
        workerId: wid,
      });
    };

    if (e.data.type === "initialize") {
      workerId = e.data.workerId;
      log(workerId, "Initializing worker", workerId);
      let compileStartTime = Date.now();
      log(workerId, "Initializing golang wasm");

      // eslint-disable-next-line no-undef
      go = new Go();

      let result = await WebAssembly.instantiate(
        e.data.module,
        go.importObject
      );

      go.run(result.instance);

      log(workerId, "Compiled in", Date.now() - compileStartTime, "ms");

      log(workerId, "Initializing context...");

      let initStartTime = Date.now();

      // Initialize rendering context
      self.initialize(e.data.initializeParams);

      log(
        workerId,
        "Context initialized in ",
        Date.now() - initStartTime,
        "ms"
      );
      renderFunc = self.render;

      // Worker manager is allowed to send render messages to this worker
      postMessage({
        initDone: true,
      });
    } else if (e.data.type === "render") {
      let renderStartTime = Date.now();

      // Main render call
      let outputRaw = renderFunc(e.data.renderParams);

      log(workerId, "Rendering complete!");
      log(workerId, "Took", Date.now() - renderStartTime, "ms");

      postMessage({
        done: true,
        workerId: workerId,
        output: JSON.parse(outputRaw),
      });
    }
  };
}
