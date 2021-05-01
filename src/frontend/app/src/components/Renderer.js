import { Row, Button, Container, Form, Col } from "react-bootstrap";

import BaseComponent from "../components/Common/BaseComponent";
import React from "react";
import RendererFrame from "./Common/RendererFrame";
import RendererStats from "./Common/RendererStats";

import MD5 from "crypto-js/md5";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

import {
  translate,
  rotateAroundXAxis,
  rotateAroundYAxis,
  rotateAroundZAxis,
  multiplyMatrices,
} from "../utility/matrix";

export default class Renderer extends BaseComponent {
  constructor(props) {
    super(props);
    this.rendererFrameRef = React.createRef();
    this.workers = {};
    this.state = {
      running: false,
      reloadOnRender: true,
      imageData: null,
      imageDetails: null,
      renderKey: 0,
      params: {
        width: 500,
        height: 500,
        scale: 100,
        x: -0.225,
        y: 2.55,
        z: 6,
        rx: 0,
        ry: 0,
        rz: 0,
        projection: 0,
        fieldOfView: 45,
        ortographicSize: 3,
        bounces: 1,
        bounceRays: 10,
        raysPerPixel: 1,
        sceneData: null,
        objData: "",
        mtlData: "",
        workerCount: 1,
        taskCount: 1,
        rngSeed: this.getNewSeed(),
      },
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
      params: { ...this.state.params, sceneData: data },
    });
  };

  loadObj = async (objFile, mtlFile) => {
    let objRequest = await fetch(objFile);
    let mtlRequest = await fetch(mtlFile);
    let objData = await objRequest.text();
    let mtlData = await mtlRequest.text();
    await this.setStateAsync({
      ...this.state,
      params: {
        ...this.state.params,
        objData: objData,
        mtlData: mtlData,
      },
    });
  };

  handleReloadOnRenderChanged = async (event) => {
    await this.setStateAsync({
      ...this.state,
      reloadOnRender: event.target.checked,
    });
  };

  handleIntParamChanged = async (event, param) => {
    let value = parseInt(event.target.value);
    let params = {
      ...this.state.params,
    };
    params[param] = value ? value : 0;
    await this.setStateAsync({
      ...this.state,
      params: params,
    });
  };

  handleFloatParamChanged = async (event, param) => {
    let value = parseFloat(event.target.value);
    let params = {
      ...this.state.params,
    };
    params[param] = value ? value : 0;
    await this.setStateAsync({
      ...this.state,
      params: params,
    });
  };

  handleProjectionChanged = async (event) => {
    console.log(this.state.params.projection, event);
    await this.setStateAsync({
      ...this.state,
      params: {
        ...this.state.params,
        projection: event.target.options.selectedIndex,
      },
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

  onRenderClicked = async (e) => {
    e.preventDefault();
    await this.wasmRender();
  };

  wasmRender = async () => {
    await this.abort();

    // Clear the view
    await this.setStateAsync({
      ...this.state,
      running: true,
      renderKey: this.state.renderKey + 1, // Re-keys the component, forces recreation
    });

    if (this.state.reloadOnRender) {
      await this.reloadWebAssembly();
    }

    let x = this.state.params.x;
    let y = this.state.params.y;
    let z = this.state.params.z;
    let rx = (this.state.params.rx * Math.PI) / 180.0;
    let ry = (this.state.params.ry * Math.PI) / 180.0;
    let rz = (this.state.params.rz * Math.PI) / 180.0;
    let cameraTransform = translate(x, y, z);
    cameraTransform = multiplyMatrices(cameraTransform, rotateAroundXAxis(rx));
    cameraTransform = multiplyMatrices(cameraTransform, rotateAroundYAxis(ry));
    cameraTransform = multiplyMatrices(cameraTransform, rotateAroundZAxis(rz));

    let initializeParams = {
      Width: this.state.params.width,
      Height: this.state.params.height,
      Camera: {
        Transform: cameraTransform,
        ProjectionPlaneDistance: 0.01,
        RaysPerPixel: this.state.params.raysPerPixel,
        Projection: this.state.params.projection,
        OrtographicSize: this.state.params.ortographicSize,
        FieldOfView: this.state.params.fieldOfView,
      },
      Settings: {
        DrawSurfaceNormal: true,
      },
      BounceLimit: this.state.params.bounces,
      BounceRays: this.state.params.bounceRays,
      Scene: {}, //this.state.params.sceneData,
      ObjBuffer: this.state.params.objData,
      MtlBuffer: this.state.params.mtlData,
      WorkerId: 0,
    };

    // Seed each task with rngSeed + taskId
    let rngSeedBase = this.state.params.rngSeed;

    this.workers = {};

    let workerIds = [...Array(this.state.params.workerCount).keys()];
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
    // Round task count to closest square
    let taskCount = this.state.params.taskCount;
    taskCount = Math.pow(Math.trunc(Math.sqrt(taskCount)), 2);
    taskCount = Math.max(1, taskCount);
    await this.setStateAsync({
      ...this.state,
      params: {
        ...this.state.params,
        taskCount: taskCount,
      },
    });

    // Generate tasks of roughly equal size. Last row/column can
    // be a few pixels larger to accommodate any resolution
    let tasksPerDimension = Math.round(Math.sqrt(taskCount));
    let taskWidth = Math.floor(this.state.params.width / tasksPerDimension);
    let taskHeight = Math.floor(this.state.params.height / tasksPerDimension);
    for (let j = 0; j < tasksPerDimension; j++) {
      for (let i = 0; i < tasksPerDimension; i++) {
        let taskId = i + j * tasksPerDimension;
        let width = taskWidth;
        let height = taskHeight;
        // Make sure the last column/row renders any remaining pixels
        // from the flooring above
        if (i == tasksPerDimension - 1) {
          width =
            taskWidth +
            (this.state.params.width - tasksPerDimension * taskWidth);
        }
        if (j == tasksPerDimension - 1) {
          height =
            taskHeight +
            (this.state.params.height - tasksPerDimension * taskHeight);
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

          if (!(key in workerEventData)) {
            workerEventData[key] = {};
            workerEventData[key].startTime = Date.now();
          }
          workerEventData[key].timer =
            Date.now() - workerEventData[key].startTime;
          workerEventData[key].progress = progress;

          let data = { ...this.state.renderEventData };
          data[workerId] = workerEventData;
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

  onAbortClicked = async () => {
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
      imageDetails: {
        width: this.state.params.width,
        height: this.state.params.height,
        scale: this.state.params.scale,
      },
    });

    // Destroy render tasks
    this.renderTasks = [];
  };

  getNewSeed = () => {
    return Math.floor(Math.random() * Number.MAX_SAFE_INTEGER);
  };

  render() {
    return (
      <Container>
        <Row>
          <h1>Renderer</h1>
        </Row>
        <Form>
          <Row>
            <Form.Group controlId="formCheckboxReloadOnRender">
              <Form.Check
                type="checkbox"
                label="Reload WebAssembly on render"
                checked={this.state.reloadOnRender}
                onChange={this.handleReloadOnRenderChanged}
              />
            </Form.Group>
          </Row>
          <Row>
            <Form.Group controlId="formWidth" className="form-margin">
              <Form.Label>Width px</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Width"
                value={this.state.params.width}
                onChange={(e) => this.handleIntParamChanged(e, "width")}
              />
            </Form.Group>

            <Form.Group controlId="formHeight" className="form-margin">
              <Form.Label>Height px</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Height"
                value={this.state.params.height}
                onChange={(e) => this.handleIntParamChanged(e, "height")}
              />
            </Form.Group>

            <Form.Group controlId="formScale" className="form-margin">
              <Form.Label>Scale %</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                value={this.state.params.scale}
                onChange={(e) => this.handleIntParamChanged(e, "scale")}
              />
            </Form.Group>

            <Form.Group controlId="formProjection" className="form-margin">
              <Form.Label>Projection</Form.Label>
              <Form.Control
                as="select"
                value={this.state.params.projection}
                onChange={this.handleProjectionChanged}
              >
                <option>Perspective</option>
                <option>Ortographic</option>
              </Form.Control>
            </Form.Group>

            <Form.Group
              hidden={this.state.params.projection !== 0}
              controlId="formFoV"
              className="form-margin"
            >
              <Form.Label>Field of View</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Height"
                value={this.state.params.fieldOfView}
                onChange={(e) => this.handleIntParamChanged(e, "fieldOfView")}
              />
            </Form.Group>

            <Form.Group
              hidden={this.state.params.projection !== 1}
              controlId="formOrtographicSize"
              className="form-margin"
            >
              <Form.Label>Ortographic Size</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Height"
                value={this.state.params.ortographicSize}
                onChange={(e) =>
                  this.handleIntParamChanged(e, "ortographicSize")
                }
              />
            </Form.Group>

            <Form.Group controlId="formRaysPerPixel" className="form-margin">
              <Form.Label>Rays Per Pixel</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Rays Per Pixel"
                value={this.state.params.raysPerPixel}
                onChange={(e) => this.handleIntParamChanged(e, "raysPerPixel")}
              />
            </Form.Group>

            <Form.Group controlId="formBounces" className="form-margin">
              <Form.Label>Bounces</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Bounces"
                value={this.state.params.bounces}
                onChange={(e) => this.handleIntParamChanged(e, "bounces")}
              />
            </Form.Group>

            <Form.Group controlId="formBounceRays" className="form-margin">
              <Form.Label>Bounce Rays</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Bounce Rays"
                value={this.state.params.bounceRays}
                onChange={(e) => this.handleIntParamChanged(e, "bounceRays")}
              />
            </Form.Group>
          </Row>
          <Row>
            <Form.Group controlId="formX" className="form-margin">
              <Form.Label>X</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                value={parseFloat(this.state.params.x)}
                onChange={(e) => this.handleFloatParamChanged(e, "x")}
              />
            </Form.Group>
            <Form.Group controlId="formY" className="form-margin">
              <Form.Label>Y</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                value={parseFloat(this.state.params.y)}
                onChange={(e) => this.handleFloatParamChanged(e, "y")}
              />
            </Form.Group>
            <Form.Group controlId="formZ" className="form-margin">
              <Form.Label>Z</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                value={parseFloat(this.state.params.z)}
                onChange={(e) => this.handleFloatParamChanged(e, "z")}
              />
            </Form.Group>

            <Form.Group controlId="formrX" className="form-margin">
              <Form.Label>rX</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                value={parseFloat(this.state.params.rx)}
                onChange={(e) => this.handleFloatParamChanged(e, "rx")}
              />
            </Form.Group>
            <Form.Group controlId="formrY" className="form-margin">
              <Form.Label>rY</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                value={parseFloat(this.state.params.ry)}
                onChange={(e) => this.handleFloatParamChanged(e, "ry")}
              />
            </Form.Group>
            <Form.Group controlId="formrZ" className="form-margin">
              <Form.Label>rZ</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                value={parseFloat(this.state.params.rz)}
                onChange={(e) => this.handleFloatParamChanged(e, "rz")}
              />
            </Form.Group>
          </Row>
          <Row>
            <Form.Group controlId="formWorkerCount" className="form-margin">
              <Form.Label>Workers</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Workers"
                value={this.state.params.workerCount}
                onChange={(e) => this.handleIntParamChanged(e, "workerCount")}
              />
            </Form.Group>
            <Form.Group controlId="formTaskCount" className="form-margin">
              <Form.Label>
                Tasks (2<sup>n</sup>)
              </Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Tasks"
                value={this.state.params.taskCount}
                onChange={(e) => this.handleIntParamChanged(e, "taskCount")}
              />
            </Form.Group>

            <Form.Group controlId="formRngSeed" className="form-margin">
              <Form.Label>RNG Seed</Form.Label>
              <Form.Control
                htmlSize="22"
                type="text"
                label="Tasks"
                value={this.state.params.rngSeed}
                onChange={(e) => this.handleIntParamChanged(e, "rngSeed")}
              />
            </Form.Group>
          </Row>
          <Row>
            <Button
              variant="primary"
              type="submit"
              className="form-margin"
              onClick={this.onRenderClicked}
            >
              Render
            </Button>
            {this.state.running ? (
              <Button variant="danger" onClick={this.onAbortClicked}>
                Abort
              </Button>
            ) : null}
          </Row>
        </Form>
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
              <Button variant="outline-primary" onClick={this.onCopyClicked}>
                <FontAwesomeIcon
                  icon={["fas", "copy"]}
                  style={{ marginRight: "0.5rem" }}
                />
                Copy
              </Button>
            </Row>
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
