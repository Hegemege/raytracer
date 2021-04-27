import { Row, Button, Container, Form } from "react-bootstrap";

import BaseComponent from "../components/Common/BaseComponent";
import React from "react";
import RendererFrame from "./Common/RendererFrame";

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
    this.worker = null;
    this.state = {
      reloadOnRender: true,
      imageData: null,
      renderKey: 0,
      params: {
        width: 500,
        height: 500,
        scale: 1.0,
        x: -0.25,
        y: 2.5,
        z: 4,
        rx: 0,
        ry: 0,
        rz: 0,
      },
      sceneData: null,
      objData: "",
      mtlData: "",
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
    if (this.worker) {
      this.worker.terminate();
    }

    this.worker = new window.Worker("go_webworker.js");

    let wasmSource = await fetch("http://localhost:8090/main.wasm");
    let data = await wasmSource.arrayBuffer();
    this.moduleData = data;
  };

  loadScene = async (sceneFile) => {
    let request = await fetch(sceneFile);
    let data = await request.json();
    await this.setStateAsync({ ...this.state, sceneData: data });
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

  handleReloadOnRenderChanged = async (event) => {
    await this.setStateAsync({
      ...this.state,
      reloadOnRender: event.target.checked,
    });
  };

  handleWidthChanged = async (event) => {
    let value = parseInt(event.target.value);
    await this.setStateAsync({
      ...this.state,
      params: {
        ...this.state.params,
        width: value ? value : 0,
      },
    });
  };

  handleHeightChanged = async (event) => {
    let value = parseInt(event.target.value);
    await this.setStateAsync({
      ...this.state,
      params: {
        ...this.state.params,
        height: value ? value : 0,
      },
    });
  };

  handleScaleChanged = async (event) => {
    let value = parseInt(event.target.value);
    await this.setStateAsync({
      ...this.state,
      params: {
        ...this.state.params,
        scale: (value ? value : 0) / 100.0,
      },
    });
  };

  handleCameraParamChanged = async (event, param) => {
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
    // Clear the view
    await this.setStateAsync({
      ...this.state,
      imageData: null,
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

    let params = {
      Width: this.state.params.width,
      Height: this.state.params.height,
      Camera: {
        Transform: cameraTransform,
        ProjectionPlaneDistance: 0.01,
        RaysPerPixel: 1,
        Projection: 0,
        OrtographicSize: 5,
        FieldOfView: 60,
      },
      Settings: {
        DrawSurfaceNormal: true,
      },
      Scene: this.state.sceneData,
      ObjBuffer: this.state.objData,
      MtlBuffer: this.state.mtlData,
    };

    // Listen to messages from the worker
    this.worker.addEventListener("message", async (event) => {
      if (event.data.message) {
        console.log("%c [WebWorker] " + event.data.message, "color: orange;");
      }

      if (event.data.done) {
        await this.setStateAsync({
          ...this.state,
          imageData: event.data.output.imageData,
        });
      }
    });

    // Start the worker
    // Each worker has to compile the source because it is not possible to
    this.worker.postMessage({
      module: this.moduleData,
      params: JSON.stringify(params),
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
            <Form.Group controlId="formWidth" className="param-field">
              <Form.Label>Width px</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Width"
                defaultValue={this.state.params.width}
                onChange={this.handleWidthChanged}
              />
            </Form.Group>

            <Form.Group controlId="formHeight" className="param-field">
              <Form.Label>Height px</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Height"
                defaultValue={this.state.params.height}
                onChange={this.handleHeightChanged}
              />
            </Form.Group>

            <Form.Group controlId="formScale" className="param-field">
              <Form.Label>Scale %</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                defaultValue={parseInt(this.state.params.scale * 100)}
                onChange={this.handleScaleChanged}
              />
            </Form.Group>
          </Row>
          <Row>
            <Form.Group controlId="formX" className="param-field">
              <Form.Label>X</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                defaultValue={parseFloat(this.state.params.x)}
                onChange={(e) => this.handleCameraParamChanged(e, "x")}
              />
            </Form.Group>
            <Form.Group controlId="formY" className="param-field">
              <Form.Label>Y</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                defaultValue={parseFloat(this.state.params.y)}
                onChange={(e) => this.handleCameraParamChanged(e, "y")}
              />
            </Form.Group>
            <Form.Group controlId="formZ" className="param-field">
              <Form.Label>Z</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                defaultValue={parseFloat(this.state.params.z)}
                onChange={(e) => this.handleCameraParamChanged(e, "z")}
              />
            </Form.Group>

            <Form.Group controlId="formrX" className="param-field">
              <Form.Label>rX</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                defaultValue={parseFloat(this.state.params.rx)}
                onChange={(e) => this.handleCameraParamChanged(e, "rx")}
              />
            </Form.Group>
            <Form.Group controlId="formrY" className="param-field">
              <Form.Label>rY</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                defaultValue={parseFloat(this.state.params.ry)}
                onChange={(e) => this.handleCameraParamChanged(e, "ry")}
              />
            </Form.Group>
            <Form.Group controlId="formrZ" className="param-field">
              <Form.Label>rZ</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                defaultValue={parseFloat(this.state.params.rz)}
                onChange={(e) => this.handleCameraParamChanged(e, "rz")}
              />
            </Form.Group>
          </Row>
          <Row>
            <Button
              variant="primary"
              type="submit"
              onClick={this.onRenderClicked}
            >
              Render
            </Button>
          </Row>
        </Form>
        <Row>
          <RendererFrame
            ref={this.rendererFrameRef}
            key={this.state.renderKey}
            imageData={this.state.imageData}
            scale={this.state.params.scale}
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
      </Container>
    );
  }
}
