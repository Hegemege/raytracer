import { Row, Button, Container, Col, Dropdown } from "react-bootstrap";

import BaseComponent from "../Common/BaseComponent";
import React from "react";
import RendererFrame from "./RendererFrame";
import RendererStats from "./RendererStats";
import RendererParams from "./RendererParams";

import MD5 from "crypto-js/md5";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

import {
  translate,
  rotateAroundXAxis,
  rotateAroundYAxis,
  rotateAroundZAxis,
  multiplyMatrices,
} from "../../utility/matrix";

import { loadFromIndexedDB, saveToIndexedDB } from "../../utility/indexeddb";

export default class Renderer extends BaseComponent {
  constructor(props) {
    super(props);
    this.rendererFrameRef = React.createRef();
    this.workers = {};
    this.state = {
      running: false,
      initialized: false,
      completed: false,
      initializeStartTime: null,
      initializeEndTime: null,
      renderStartTime: null,
      renderEndTime: null,
      reloadOnRender: true,
      imageData: null,
      imageDetails: null,
      renderKey: 0,
      sceneData: null,
      objData: "",
      mtlData: "",
      renderEventData: {},
      renderTasks: [],
      renderParams: null,
      selectedPreset: 0,
      loadingPreset: false,
    };
    this.textureData = [];
    this.renderTasks = [];
    this.taskIdBase = 0;
    this.presets = [];
    this.forceReload = false;
  }

  getPreset = () => {
    return this.presets[this.state.selectedPreset];
  };

  componentDidMount = async () => {
    await this.reloadWebAssembly();

    // Load presets
    await this.loadPreset("presets/cornell-box.json");
    await this.loadPreset("presets/conference.json");
    await this.loadPreset("presets/sponza.json");
    await this.loadPreset("presets/sibenik.json");

    // TODO: Load saved presets from localstorage?

    let preset = this.getPreset();
    this.setPreset(preset);
  };

  loadPreset = async (path) => {
    let preset = await (await fetch(path)).json();
    preset.key = Math.random();
    this.presets.push(preset);
  };

  handlePresetChange = async (e, preset, index) => {
    // If the index didn't change, no need to re-initialize the workers
    let samePreset = this.state.selectedPreset === index;
    preset.key = Math.random();

    await this.setStateAsync({
      ...this.state,
      loadingPreset: true,
      initialized: samePreset ? this.state.initialized : false,
      completed: samePreset ? this.state.completed : false,
      renderEventData: samePreset ? this.state.renderEventData : {},
      imageData: samePreset ? this.state.imageData : [],
      renderKey: samePreset ? this.state.renderKey : this.state.renderKey + 1,
    });

    await this.setPreset(preset);

    await this.setStateAsync({
      ...this.state,
      selectedPreset: index,
      loadingPreset: false,
    });
  };

  setPreset = async (preset) => {
    await this.loadObj(
      preset.params.objectPath,
      preset.params.materialPath,
      preset.params.texturePaths
    );
  };

  reloadWebAssembly = async () => {
    let wasmSource = await fetch("http://localhost:8090/main.wasm");
    let data = await wasmSource.arrayBuffer();
    this.moduleData = data;
  };

  loadScene = async (sceneFile) => {
    let request = await fetch(sceneFile);
    let data = await request.json();
    await this.setStateAsync({
      ...this.state,
      sceneData: data,
    });
  };

  loadObj = async (objFile, mtlFile, textureFiles) => {
    let objRequest = await fetch(objFile);
    let mtlRequest = await fetch(mtlFile);
    let objData = await objRequest.text();
    let mtlData = await mtlRequest.text();

    let textures = [];
    for (let texture of textureFiles) {
      let textureRequest = await fetch(texture.path);
      let textureData = await textureRequest.arrayBuffer();
      textures.push({
        Name: texture.name,
        Buffer: textureData, //this.arrayBufferToBase64(textureData),
      });
    }

    await this.setStateAsync({
      ...this.state,
      objData: objData,
      mtlData: mtlData,
    });

    this.textureData = textures;
  };

  getWasmMd5 = (arrayBuffer) => {
    // do something with the text response
    let hex = [...new Uint8Array(arrayBuffer)]
      .map((x) => x.toString(16).padStart(2, "0"))
      .join("");

    let md5 = MD5(hex);
    return md5.toString();
  };

  onStartRender = async (e, params) => {
    await this.wasmRender(params);
  };

  onInitializeContext = async (e, params) => {
    await this.reloadWebAssembly();
    await this.wasmSetup(params);
  };

  wasmSetup = async (params) => {
    await this.abort();

    // Clear the view
    await this.setStateAsync({
      ...this.state,
      running: true,
      rendering: false,
      completed: false,
      initialized: false,
      initializeStartTime: Date.now(),
      initializeEndTime: null,
      renderStartTime: null,
      renderEndTime: null,
      aborted: false,
      renderKey: this.state.renderKey + 1, // Re-keys the component, forces recreation
    });

    if (this.state.reloadOnRender) {
      await this.reloadWebAssembly();
    }

    let initializeParams = {
      Debug: false,
      UseBVH: params.useBVH,
      BVHMaxLeafSize: params.maxLeafSize,
      BVHMaxDepth: params.maxDepth,
      Scene: {}, //this.state.sceneData,
      ObjBuffer: this.state.objData,
      MtlBuffer: this.state.mtlData,
      RawTextures: this.textureData.map((texture) => {
        return {
          Name: texture.Name,
        };
      }),
      WorkerId: 0,
    };

    this.workers = {};
    this.taskIdBase = 0;

    let workerIds = [...Array(params.workerCount).keys()];
    let setupPromises = [];
    for (let workerId of workerIds) {
      let workerPromise = this.setupWorker(
        workerId,
        JSON.parse(JSON.stringify(initializeParams))
      );
      setupPromises.push(workerPromise);
    }

    // Wait for initialization
    let workers = await Promise.all(setupPromises);

    for (let worker of workers) {
      this.workers[worker.workerId] = worker;
    }

    if (this.state.aborted) {
      await this.terminateWorkers();
      return;
    }

    let bvhKey = params.objectPath;
    let bvhData = undefined;

    if (params.loadBVH) {
      try {
        let savedBVH = await loadFromIndexedDB("bvhStore", bvhKey);
        bvhData = savedBVH.data;
      } catch (error) {
        // No need to do anything
      }
    }

    if (!bvhData) {
      // Build BVH
      let bvhWorker = workers[0];
      let buildBVHPromise = new Promise((resolve) => {
        bvhWorker.worker.addEventListener("message", async (event) => {
          if (event.data.buildBVHDone) {
            resolve(event.data.output);
          }
        });

        this.buildBVHWorker(bvhWorker);
      });

      bvhData = JSON.stringify(await buildBVHPromise);
    }

    if (params.saveBVH) {
      let rawData = {
        id: bvhKey,
        data: bvhData,
      };

      try {
        await saveToIndexedDB("bvhStore", rawData);
      } catch (error) {
        console.log("Could not save to indexed DB", error);
      }
    }

    if (this.state.aborted) {
      await this.terminateWorkers();
      return;
    }

    // Load BVH to all workers
    let loadBVHPromises = [];
    for (let worker of workers) {
      loadBVHPromises.push(
        new Promise((resolve) => {
          worker.worker.addEventListener("message", async (event) => {
            if (event.data.loadBVHDone) {
              resolve();
            }
          });

          this.loadBVHWorker(worker, bvhData);
        })
      );
    }

    await Promise.all(loadBVHPromises);

    if (this.state.aborted) {
      await this.terminateWorkers();
      return;
    }

    await this.setStateAsync({
      ...this.state,
      initialized: true,
      initializeEndTime: Date.now(),
    });

    if (params.renderAfterInitialization) {
      await this.wasmRender(params);
    }
  };

  wasmRender = async (params) => {
    await this.setStateAsync({
      ...this.state,
      imageData: [],
      running: true,
      completed: false,
      renderEndTime: null,
      rendering: true,
      renderStartTime: Date.now(),
      imageDetails: {
        width: params.width,
        height: params.height,
        scale: params.scale,
      },
      renderKey: this.state.renderKey + 1, // Re-keys the component, forces recreation
    });

    // Clean any spawn rays, trace or output event data
    for (let workerId in this.state.renderEventData) {
      let workerEventData = this.state.renderEventData[workerId];

      for (let taskKey in workerEventData) {
        if (
          taskKey.includes("spawnRays") ||
          taskKey.includes("trace") ||
          taskKey.includes("output")
        ) {
          delete workerEventData[taskKey];
        }
      }

      let data = { ...this.state.renderEventData };
      data[workerId] = workerEventData;
      await this.setStateAsync({
        ...this.state,
        renderEventData: data,
      });
    }

    this.renderTasks = [];

    // Seed each task with rngSeed + taskId
    let rngSeedBase = params.rngSeed;

    // Generate render tasks
    let tasksPerDimension = params.taskCount;
    tasksPerDimension = tasksPerDimension ? tasksPerDimension : 1;
    tasksPerDimension = Math.trunc(Math.sqrt(tasksPerDimension));
    tasksPerDimension = Math.max(1, tasksPerDimension);

    let x = parseFloat(params.x);
    let y = parseFloat(params.y);
    let z = parseFloat(params.z);
    let rx = (parseFloat(params.rx) * Math.PI) / 180.0;
    let ry = (parseFloat(params.ry) * Math.PI) / 180.0;
    let rz = (parseFloat(params.rz) * Math.PI) / 180.0;
    let cameraTransform = translate(x, y, z);
    cameraTransform = multiplyMatrices(cameraTransform, rotateAroundXAxis(rx));
    cameraTransform = multiplyMatrices(cameraTransform, rotateAroundYAxis(ry));
    cameraTransform = multiplyMatrices(cameraTransform, rotateAroundZAxis(rz));

    // Generate tasks of roughly equal size. Last row/column can
    // be a few pixels larger to accommodate any resolution
    let taskWidth = Math.floor(params.width / tasksPerDimension);
    let taskHeight = Math.floor(params.height / tasksPerDimension);
    for (let j = 0; j < tasksPerDimension; j++) {
      for (let i = 0; i < tasksPerDimension; i++) {
        let taskId = this.taskIdBase + i + j * tasksPerDimension;
        let width = taskWidth;
        let height = taskHeight;
        // Make sure the last column/row renders any remaining pixels
        // from the flooring above
        if (i == tasksPerDimension - 1) {
          width = taskWidth + (params.width - tasksPerDimension * taskWidth);
        }
        if (j == tasksPerDimension - 1) {
          height =
            taskHeight + (params.height - tasksPerDimension * taskHeight);
        }
        let task = {
          Camera: {
            Transform: cameraTransform,
            ProjectionPlaneDistance: 0.01,
            RaysPerPixel: params.raysPerPixel,
            Projection: params.projection,
            OrtographicSize: params.ortographicSize,
            FieldOfView: params.fieldOfView,
          },
          TotalWidth: params.width,
          TotalHeight: params.height,
          TaskID: taskId,
          RenderKey: this.state.renderKey,
          Width: width,
          Height: height,
          XOffset: i * taskWidth,
          YOffset: j * taskHeight,
          RNGSeed: rngSeedBase + taskId,
          Settings: {
            DrawSurfaceNormal: true,
            GammaCorrection: params.gammaCorrection,
            Gamma: parseFloat(params.gamma),
            BounceLimit: params.bounces,
            LightSampleRays: params.lightSampleRays,
            LightIntensity: parseFloat(params.lightIntensity),
            DebugLightSize: parseFloat(params.debugLightSize),
          },
          IncrementalRendering: params.incrementalRendering,
        };

        this.renderTasks.push(task);
      }
    }

    this.taskIdBase += tasksPerDimension * tasksPerDimension;

    for (let workerId in this.workers) {
      let worker = this.workers[workerId];
      worker.done = false;
      worker.output = null;

      let type = "askForWork";
      if (params.incrementalRendering) {
        type = "askForWorkIncremental";
      }

      worker.worker.postMessage({
        workerId: worker.workerId,
        type: type,
      });
    }
  };

  setupWorker = async (workerId, initializeParams) => {
    let workerScript = new window.Worker("go_webworker.js");
    let worker = {
      workerId: workerId,
      worker: workerScript,
      done: false,
      output: null,
      initialized: false,
    };

    let rawParams = JSON.stringify(initializeParams);

    return new Promise((resolve) => {
      worker.worker.addEventListener("message", async (event) => {
        this.workerLogger(event);
      });
      worker.worker.addEventListener("message", async (event) => {
        this.workerProgressUpdate(event, worker);
      });
      worker.worker.addEventListener("message", async (event) => {
        this.workerRenderDone(event, worker);
      });
      worker.worker.addEventListener("message", async (event) => {
        this.workerIncrementalRenderDone(event, worker);
      });

      // Listen to messages from the worker
      worker.worker.addEventListener("message", async (event) => {
        // Messages from the WebWorker JS side
        if (event.data.initDone) {
          worker.initialized = true;
          resolve(worker);
        }
      });

      this.initializeWorker(worker, rawParams);
    });
  };

  workerLogger = async (event) => {
    if (event.data.logMessage) {
      console.log(
        "%c [WebWorker " +
          event.data.workerId.toString() +
          "] " +
          event.data.message,
        "color: orange;"
      );
      return;
    }
  };

  workerProgressUpdate = async (event, worker) => {
    if (event.data.progressUpdate) {
      let workerEventData = {};
      if (worker.workerId in this.state.renderEventData) {
        workerEventData = this.state.renderEventData[worker.workerId];
      }

      let key = event.data.data.event;
      let progress = event.data.data.progress;
      let taskId = event.data.data.taskId;
      let rays = event.data.data.rays;

      if (!(key in workerEventData)) {
        workerEventData[key] = {};
      }
      if (!(taskId in workerEventData[key])) {
        workerEventData[key][taskId] = {};
        workerEventData[key][taskId].startTime = Date.now();
      }

      workerEventData[key][taskId].timer =
        Date.now() - workerEventData[key][taskId].startTime;
      workerEventData[key][taskId].progress = progress;

      workerEventData.rays = rays;

      let data = { ...this.state.renderEventData };
      data[worker.workerId] = workerEventData;
      await this.setStateAsync({
        ...this.state,
        renderEventData: data,
      });
    }
  };

  workerRenderDone = async (event, worker) => {
    if (event.data.renderDone) {
      // Completion of the WebWorker
      if (event.data.output) {
        let params = JSON.parse(event.data.params);
        let imageData = [...this.state.imageData];
        imageData.push({
          params: params,
          imageData: event.data.output.imageData,
        });
        await this.setStateAsync({
          ...this.state,
          imageData: imageData,
        });
      }

      // Take a new task
      if (this.renderTasks.length > 0) {
        let task = this.renderTasks.pop();
        this.renderWorker(worker, task);
      } else {
        worker.done = true;

        // If all workers are done, mark no longer running
        if (
          Object.keys(this.workers).filter(
            (worker) => !this.workers[worker].done
          ).length == 0
        ) {
          await this.setStateAsync({
            ...this.state,
            running: false,
            completed: true,
            renderEndTime: Date.now(),
          });
        }
      }
    }
  };

  workerIncrementalRenderDone = async (event, worker) => {
    if (event.data.incrementalRenderPartial) {
      if (event.data.output) {
        let params = JSON.parse(event.data.params);
        let imageData = [...this.state.imageData];
        // Remove existing image data for the given worker
        let existingIndex = imageData.findIndex(
          (item) => item.params.XOffset == event.data.params.XOffset
        );
        if (existingIndex !== -1) {
          imageData.splice(existingIndex, 1);
        }
        imageData.push({
          params: params,
          imageData: event.data.output.imageData,
        });
        await this.setStateAsync({
          ...this.state,
          imageData: imageData,
        });
      } else {
        // Just starting the incremental rendering
        if (this.renderTasks.length > 0) {
          let task = this.renderTasks.pop();
          this.incrementalRenderWorker(worker, task);
        } else {
          worker.done = true;
        }
      }
    }

    if (event.data.incrementalRenderDone) {
      worker.done = true;

      // If all workers are done, mark no longer running
      if (
        Object.keys(this.workers).filter((worker) => !this.workers[worker].done)
          .length == 0
      ) {
        await this.setStateAsync({
          ...this.state,
          running: false,
          completed: true,
          renderEndTime: Date.now(),
        });
      }
    }
  };

  initializeWorker = async (worker, params) => {
    // Split texture data from params
    let textureData = [];

    for (let texture of this.textureData) {
      textureData.push(texture.Buffer.slice(0));
    }

    // Start the worker
    // Each worker has to compile the source because it is not possible to
    worker.worker.postMessage(
      {
        workerId: worker.workerId,
        module: this.moduleData,
        type: "initialize",
        initializeParams: params,
        textureData: textureData,
      },
      textureData
    );
  };

  buildBVHWorker = async (worker) => {
    // Start the worker
    // Each worker has to compile the source because it is not possible to
    worker.worker.postMessage({
      workerId: worker.workerId,
      type: "buildBVH",
    });
  };

  loadBVHWorker = async (worker, bvhData) => {
    // Start the worker
    // Each worker has to compile the source because it is not possible to
    worker.worker.postMessage({
      workerId: worker.workerId,
      type: "loadBVH",
      bvhData: bvhData,
    });
  };

  renderWorker = async (worker, task) => {
    worker.worker.postMessage({
      workerId: worker.workerId,
      taskId: task.TaskID,
      type: "render",
      renderParams: JSON.stringify(task),
    });
  };

  incrementalRenderWorker = async (worker, task) => {
    worker.worker.postMessage({
      workerId: worker.workerId,
      taskId: task.TaskID,
      type: "incrementalRender",
      renderParams: JSON.stringify(task),
      raysPerPixel: task.Camera.RaysPerPixel,
    });
  };

  onCopyClicked = async () => {
    // TODO: Move the RendererFrame with optional hiding?
    this.rendererFrameRef.current.canvasRef.current.toBlob(async (blob) => {
      // eslint-disable-next-line no-undef
      const item = new ClipboardItem({ "image/png": blob });
      navigator.clipboard.write([item]);
    });
  };

  onAbort = async () => {
    await this.abort();
  };

  terminateWorkers = async () => {
    for (let key in this.workers) {
      this.workers[key].worker.terminate();
    }
  };

  abort = async () => {
    // Destroy existing workers
    await this.terminateWorkers();

    // Destroy event data
    await this.setStateAsync({
      ...this.state,
      renderEventData: {},
      imageData: [],
      running: false,
      rendering: false,
      initialized: false,
      completed: true,
      aborted: true,
      initializeEndTime: this.state.initialized
        ? this.state.initializeEndTime
        : Date.now(),
      renderStartTime: this.state.initialized
        ? this.state.renderStartTime
        : Date.now(),
      renderEndTime: Date.now(),
    });

    // Destroy render tasks
    this.renderTasks = [];
  };

  onParamsChanged = async (params) => {
    await this.setStateAsync({
      ...this.state,
      renderParams: params,
    });
  };

  arrayBufferToBase64 = (buffer) => {
    var binary = "";
    var bytes = new Uint8Array(buffer);
    var len = bytes.byteLength;
    for (var i = 0; i < len; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary);
  };

  render() {
    let initializeEndTime = this.state.initialized
      ? this.state.initializeEndTime
      : Date.now();
    let renderEndTime = this.state.completed
      ? this.state.renderEndTime
      : Date.now();

    let totalRays = 0;
    let raysPerSecond = 0;
    if (this.state.rendering || this.state.completed) {
      for (let key in this.state.renderEventData) {
        totalRays += this.state.renderEventData[key].rays;
      }

      let renderTotalSeconds =
        (renderEndTime - this.state.renderStartTime) / 1000.0;
      raysPerSecond = totalRays / renderTotalSeconds;
    }

    let initializeTime = this.state.initializeStartTime
      ? initializeEndTime - this.state.initializeStartTime
      : 0;
    let renderTime = this.state.renderStartTime
      ? renderEndTime - this.state.renderStartTime
      : 0;
    let totalTime = initializeTime + renderTime;

    let estimatedRays = 0;
    let estimatedTimeRemaining = 0;
    if (this.state.renderParams) {
      estimatedRays =
        this.state.renderParams.width *
        this.state.renderParams.height *
        this.state.renderParams.raysPerPixel *
        (1 +
          this.state.renderParams.lightSampleRays +
          this.state.renderParams.lightSampleRays *
            this.state.renderParams.bounces);

      if (this.state.running) {
        let remainingRays = estimatedRays - totalRays;
        estimatedTimeRemaining = remainingRays / raysPerSecond;
      }
    }

    // Presets
    let preset = this.presets[this.state.selectedPreset];
    let chosenPresetName = preset ? preset.name : "No preset chosen";

    return (
      <Container>
        <Row>
          <h1>Renderer</h1>
        </Row>
        <Row className="preset-selector">
          <span>Preset</span>
          <Dropdown className="left-margin">
            <Dropdown.Toggle
              variant="outline-primary"
              size="sm"
              disabled={this.state.running || this.state.loadingPreset}
            >
              {chosenPresetName}
            </Dropdown.Toggle>

            <Dropdown.Menu>
              {this.presets.map((preset, i) => (
                <Dropdown.Item
                  key={i}
                  onClick={(e) => this.handlePresetChange(e, preset, i)}
                >
                  {preset.name}
                </Dropdown.Item>
              ))}
            </Dropdown.Menu>
          </Dropdown>
        </Row>
        {preset && (
          <RendererParams
            running={this.state.running}
            initialized={this.state.initialized}
            aborted={this.state.aborted}
            params={preset.params}
            paramKey={preset.key}
            onAbort={this.onAbort}
            onStartRender={this.onStartRender}
            onInitializeContext={this.onInitializeContext}
            onChanged={this.onParamsChanged}
          ></RendererParams>
        )}
        <Row>
          <Col>
            <Row>
              <RendererFrame
                ref={this.rendererFrameRef}
                key={this.state.renderKey}
                imageData={this.state.imageData}
                imageDetails={this.state.imageDetails}
              ></RendererFrame>
            </Row>
            <Row>
              <Button
                variant="outline-primary"
                className="right-margin"
                onClick={this.onCopyClicked}
              >
                <FontAwesomeIcon
                  icon={["fas", "copy"]}
                  style={{ marginRight: "0.5rem" }}
                />
                Copy
              </Button>
            </Row>
            {this.state.running || this.state.completed ? (
              <div>
                <div>Total time {totalTime} ms</div>
                <div>Initialization time {initializeTime} ms</div>
                {this.state.initialized && (
                  <div>Render time {renderTime} ms</div>
                )}
                <div>Total rays {(totalRays / 1000000).toFixed(2)}M</div>
                <div>MRays/s {(raysPerSecond / 1000000).toFixed(2)}</div>
                <div>
                  Est. time remaining {estimatedTimeRemaining.toFixed(2)} s
                </div>
              </div>
            ) : null}
            <div>
              <div>Est. rays {(estimatedRays / 1000000).toFixed(2)}M</div>
            </div>
          </Col>
          <Col>
            {Object.keys(this.state.renderEventData).map((workerId, i) => (
              <RendererStats
                key={i}
                workerId={workerId}
                data={this.state.renderEventData[workerId]}
              ></RendererStats>
            ))}
          </Col>
        </Row>
      </Container>
    );
  }
}
