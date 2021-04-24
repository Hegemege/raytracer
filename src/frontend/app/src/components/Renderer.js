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
    //console.log("WASM MD5:", this.getWasmMd5(data));

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

  getWasmMd5 = (arrayBuffer) => {
    // do something with the text response
    let hex = [...new Uint8Array(arrayBuffer)]
      .map((x) => x.toString(16).padStart(2, "0"))
      .join("");

    let md5 = MD5(hex);
    return md5.toString();
  };

  onRenderClicked = async () => {
    if (this.state.reloadOnRender) {
      await this.reloadWebAssembly();
    }

    let params = {
      width: 500,
      height: 500,
    };

    let outputRaw = window.render(JSON.stringify(params)); // Exposed from golang
    let output = JSON.parse(outputRaw);

    console.log(output);

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
        <Row>
          <Form.Group controlId="formBasicCheckboxReloadOnRender">
            <Form.Check
              type="checkbox"
              label="Reload WebAssembly on render"
              checked={this.state.reloadOnRender}
              onChange={this.handleReloadOnRenderChanged}
            />
          </Form.Group>
        </Row>
        <Row>
          <Button variant="primary" onClick={this.onRenderClicked}>
            Render
          </Button>
        </Row>
        <Row>
          <RendererFrame imageData={this.state.imageData}></RendererFrame>
        </Row>
      </Container>
    );
  }
}
