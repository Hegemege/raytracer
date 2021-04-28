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

addEventListener(
  "message",
  async (e) => {
    // Sending logging events to main thread
    let log = (line) => {
      postMessage({
        logMessage: true,
        message: line,
      });
    };

    let compileStartTime = Date.now();
    log("Initializing golang wasm");

    // eslint-disable-next-line no-undef
    let go = new Go();

    let result = await WebAssembly.instantiate(e.data.module, go.importObject);

    go.run(result.instance);

    let compileEndTime = Date.now();
    log(
      "Compiled in " + (compileEndTime - compileStartTime).toString() + " ms"
    );

    log("Rendering...");
    let startTime = Date.now();

    // Main render call
    let outputRaw = self.render(e.data.params);

    let endTime = Date.now();
    log("Rendering complete!");
    log("Took " + (endTime - startTime).toString() + " ms");

    postMessage({
      done: true,
      output: JSON.parse(outputRaw),
    });
  },
  false
);
