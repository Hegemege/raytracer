import { Row, Button, Container, Col } from "react-bootstrap";

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

export default class Renderer extends BaseComponent {
  constructor(props) {
    super(props);
    this.rendererFrameRef = React.createRef();
    this.workers = {};
    this.state = {
      running: false,
      completed: false,
      startTime: null,
      endTime: null,
      reloadOnRender: true,
      imageData: null,
      imageDetails: null,
      renderKey: 0,
      sceneData: null,
      objData: "",
      mtlData: "",
      renderEventData: {},
      renderTasks: [],
    };
  }

  componentDidMount = async () => {
    await this.reloadWebAssembly();
    await this.loadScene("scenes/simple-spheres.json");
    await this.loadObj(
      "scenes/obj/cornell-box/cornell-box.obj",
      "scenes/obj/cornell-box/cornell-box.mtl"
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

  loadObj = async (objFile, mtlFile) => {
    let objRequest = await fetch(objFile);
    let mtlRequest = await fetch(mtlFile);
    let objData = await objRequest.text();
    let mtlData = await mtlRequest.text();
    await this.setStateAsync({
      ...this.state,
      objData: objData,
      mtlData: mtlData,
    });
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

  wasmRender = async (params) => {
    await this.abort();

    // Clear the view
    await this.setStateAsync({
      ...this.state,
      running: true,
      completed: false,
      startTime: Date.now(),
      endTime: null,
      imageDetails: {
        width: params.width,
        height: params.height,
        scale: params.scale,
      },
      renderKey: this.state.renderKey + 1, // Re-keys the component, forces recreation
    });

    if (this.state.reloadOnRender) {
      await this.reloadWebAssembly();
    }

    let x = params.x;
    let y = params.y;
    let z = params.z;
    let rx = (params.rx * Math.PI) / 180.0;
    let ry = (params.ry * Math.PI) / 180.0;
    let rz = (params.rz * Math.PI) / 180.0;
    let cameraTransform = translate(x, y, z);
    cameraTransform = multiplyMatrices(cameraTransform, rotateAroundXAxis(rx));
    cameraTransform = multiplyMatrices(cameraTransform, rotateAroundYAxis(ry));
    cameraTransform = multiplyMatrices(cameraTransform, rotateAroundZAxis(rz));

    let initializeParams = {
      Width: params.width,
      Height: params.height,
      Camera: {
        Transform: cameraTransform,
        ProjectionPlaneDistance: 0.01,
        RaysPerPixel: params.raysPerPixel,
        Projection: params.projection,
        OrtographicSize: params.ortographicSize,
        FieldOfView: params.fieldOfView,
      },
      Settings: {
        DrawSurfaceNormal: true,
        Debug: false,
      },
      BounceLimit: params.bounces,
      BounceRays: params.bounceRays,
      Scene: {}, //this.state.sceneData,
      ObjBuffer: this.state.objData,
      MtlBuffer: this.state.mtlData,
      WorkerId: 0,
    };

    // Seed each task with rngSeed + taskId
    let rngSeedBase = params.rngSeed;

    this.workers = {};

    let workerIds = [...Array(params.workerCount).keys()];
    let workerPromises = [];
    for (let workerId of workerIds) {
      let workerPromise = this.setupWorker(workerId, initializeParams);
      workerPromises.push(workerPromise);
    }

    // Wait for initialization
    let workers = await Promise.all(workerPromises);

    for (let worker of workers) {
      this.workers[worker.workerId] = worker;
    }

    // Generate render tasks
    let tasksPerDimension = params.taskCount;
    tasksPerDimension = tasksPerDimension ? tasksPerDimension : 1;
    tasksPerDimension = Math.trunc(Math.sqrt(tasksPerDimension));
    tasksPerDimension = Math.max(1, tasksPerDimension);

    // Generate tasks of roughly equal size. Last row/column can
    // be a few pixels larger to accommodate any resolution
    let taskWidth = Math.floor(params.width / tasksPerDimension);
    let taskHeight = Math.floor(params.height / tasksPerDimension);
    for (let j = 0; j < tasksPerDimension; j++) {
      for (let i = 0; i < tasksPerDimension; i++) {
        let taskId = i + j * tasksPerDimension;
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
          TaskID: taskId,
          Width: width,
          Height: height,
          XOffset: i * taskWidth,
          YOffset: j * taskHeight,
          RNGSeed: rngSeedBase + taskId,
        };

        this.renderTasks.push(task);
      }
    }

    for (let worker of workers) {
      worker.worker.postMessage({
        workerId: worker.workerId,
        type: "askForWork",
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

    return new Promise((resolve) => {
      // Listen to messages from the worker
      workerScript.addEventListener("message", async (event) => {
        // Messages from the WebWorker JS side
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

        if (event.data.progressUpdate) {
          let workerEventData = {};
          if (worker.workerId in this.state.renderEventData) {
            workerEventData = this.state.renderEventData[worker.workerId];
          }

          let key = event.data.data.event;
          let progress = event.data.data.progress;
          let taskId = event.data.data.taskId;

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

          let data = { ...this.state.renderEventData };
          data[worker.workerId] = workerEventData;
          await this.setStateAsync({
            ...this.state,
            renderEventData: data,
          });
        }

        if (event.data.done) {
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
            // No tasks left, terminate
            worker.worker.terminate();
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
                endTime: Date.now(),
              });
            }
          }
        }

        if (event.data.initDone) {
          worker.initialized = true;
          resolve(worker);
        }
      });

      this.initializeWorker(worker, initializeParams);
    });
  };

  initializeWorker = async (worker, params) => {
    // Start the worker
    // Each worker has to compile the source because it is not possible to
    worker.worker.postMessage({
      workerId: worker.workerId,
      module: this.moduleData,
      type: "initialize",
      initializeParams: JSON.stringify(params),
    });
  };

  renderWorker = async (worker, params) => {
    worker.worker.postMessage({
      workerId: worker.workerId,
      taskId: params.TaskID,
      type: "render",
      renderParams: JSON.stringify(params),
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

  abort = async () => {
    // Destroy existing workers
    for (let key in this.workers) {
      this.workers[key].worker.terminate();
    }

    // Destroy event data
    await this.setStateAsync({
      ...this.state,
      renderEventData: {},
      imageData: [],
      running: false,
      completed: true,
      endTime: Date.now(),
    });

    // Destroy render tasks
    this.renderTasks = [];
  };

  onParamsChanged = async () => {};

  render() {
    let endTime = this.state.completed ? this.state.endTime : Date.now();
    return (
      <Container>
        <Row>
          <h1>Renderer</h1>
        </Row>
        <RendererParams
          running={this.state.running}
          onAbort={this.onAbort}
          onStartRender={this.onStartRender}
          onChanged={this.onParamsChanged}
        ></RendererParams>
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
                className="form-margin"
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
              <Row>
                <p>Render time {endTime - this.state.startTime} ms</p>
              </Row>
            ) : null}
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