import { Row, Button, Container, Form } from "react-bootstrap";

import BaseComponent from "../components/Common/BaseComponent";
import React from "react";
import RendererFrame from "./Common/RendererFrame";

import MD5 from "crypto-js/md5";

export default class Renderer extends BaseComponent {
  constructor(props) {
    super(props);
    this.state = {
      reloadOnRender: true,
      imageData: null,
      renderKey: 0,
      params: {
        width: 500,
        height: 500,
      },
    };
  }

  componentDidMount = async () => {
    await this.reloadWebAssembly();
  };

  reloadWebAssembly = async () => {
    window.render = undefined;

    let wasmSource = await fetch("http://localhost:8090/main.wasm");
    let data = await wasmSource.arrayBuffer();

    // Slow, but useful for testing
    // console.log("WASM MD5:", this.getWasmMd5(data));

    let result = await WebAssembly.instantiate(
      data,
      window.gowasm.importObject
    );

    window.gowasm.run(result.instance);
  };

  handleReloadOnRenderChanged = async (event) => {
    await this.setStateAsync({
      ...this.state,
      reloadOnRender: event.target.checked,
    });
  };

  handleWidthChanged = async (event) => {
    await this.setStateAsync({
      ...this.state,
      params: {
        ...this.state.params,
        width: parseInt(event.target.value),
      },
    });
  };

  handleHeightChanged = async (event) => {
    await this.setStateAsync({
      ...this.state,
      params: {
        ...this.state.params,
        height: parseInt(event.target.value),
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
    // Clear the view
    await this.setStateAsync({
      ...this.state,
      imageData: null,
      renderKey: this.state.renderKey + 1, // Re-keys the component, forces recreation
    });

    if (this.state.reloadOnRender) {
      await this.reloadWebAssembly();
    }

    let params = {
      Width: this.state.params.width,
      Height: this.state.params.height,
      CameraSettings: {
        ProjectionPlaneDistance: 0.1,
        RaysPerPixel: 1,
        Projection: 0,
        FieldOfView: 60,
      },
    };

    let outputRaw = window.render(JSON.stringify(params)); // Exposed from golang
    let output = JSON.parse(outputRaw);

    await this.setStateAsync({
      ...this.state,
      imageData: output.imageData,
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
            <Form.Group controlId="formWidth">
              <Form.Label>Width</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Width"
                defaultValue={this.state.params.width}
                onChange={this.handleWidthChanged}
              />
            </Form.Group>

            <Form.Group controlId="formHeight">
              <Form.Label>Height</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Height"
                defaultValue={this.state.params.height}
                onChange={this.handleHeightChanged}
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
            key={this.state.renderKey}
            imageData={this.state.imageData}
          ></RendererFrame>
        </Row>
      </Container>
    );
  }
}
