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
  let incrementalRenderFunc = null;
  let initializeIncrementalRenderFunc = null;
  let buildBVHFunc = null;
  let loadBVHFunc = null;

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

      let textureData = e.data.textureData.map(
        (buffer) => new Uint8Array(buffer)
      );

      // Initialize rendering context
      self.initialize(e.data.initializeParams, ...textureData);

      log(
        workerId,
        "Context initialized in ",
        Date.now() - initStartTime,
        "ms"
      );
      renderFunc = self.render;
      incrementalRenderFunc = self.incrementalRender;
      initializeIncrementalRenderFunc = self.initializeIncrementalRender;
      buildBVHFunc = self.buildBVH;
      loadBVHFunc = self.loadBVH;

      // Worker manager is allowed to send render messages to this worker
      postMessage({
        initDone: true,
      });
    } else if (e.data.type === "buildBVH") {
      log(workerId, "Building BVH");
      let buildBVHStartTime = Date.now();

      let outputRaw = buildBVHFunc();

      log(workerId, "Building BVH complete!");
      log(workerId, "Took", Date.now() - buildBVHStartTime, "ms");

      postMessage({
        buildBVHDone: true,
        workerId: workerId,
        output: JSON.parse(outputRaw),
      });
    } else if (e.data.type === "loadBVH") {
      log(workerId, "Loading BVH");
      let loadBVHStartTime = Date.now();

      loadBVHFunc(e.data.bvhData);

      log(workerId, "Loading BVH complete!");
      log(workerId, "Took", Date.now() - loadBVHStartTime, "ms");

      postMessage({
        loadBVHDone: true,
        workerId: workerId,
      });
    } else if (e.data.type === "render") {
      log(workerId, "Rendering task", e.data.taskId);
      let renderStartTime = Date.now();

      // Main render call
      let outputRaw = renderFunc(e.data.renderParams);

      log(workerId, "Rendering complete!");
      log(workerId, "Took", Date.now() - renderStartTime, "ms");

      postMessage({
        renderDone: true,
        workerId: workerId,
        output: JSON.parse(outputRaw),
        params: e.data.renderParams, // return original params for parsing the final image
      });
    } else if (e.data.type === "incrementalRender") {
      log(workerId, "Incremental rendering task", e.data.taskId);
      let renderStartTime = Date.now();

      initializeIncrementalRenderFunc(e.data.renderParams);

      for (let i = 0; i < e.data.raysPerPixel; i++) {
        // Main render call
        let outputRaw = incrementalRenderFunc(e.data.renderParams);

        postMessage({
          incrementalRenderPartial: true,
          workerId: workerId,
          output: JSON.parse(outputRaw),
          params: e.data.renderParams, // return original params for parsing the final image
        });
      }

      postMessage({
        incrementalRenderDone: true,
        workerId: workerId,
      });

      log(workerId, "Rendering complete!");
      log(workerId, "Took", Date.now() - renderStartTime, "ms");
    } else if (e.data.type === "askForWork") {
      // Used for the first task
      postMessage({
        renderDone: true,
        workerId: workerId,
        output: null,
      });
    } else if (e.data.type === "askForWorkIncremental") {
      // Used for the first task
      postMessage({
        incrementalRenderPartial: true,
        workerId: workerId,
        output: null,
      });
    }
  };
}
