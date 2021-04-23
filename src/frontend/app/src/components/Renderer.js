import { Row, Button, Container, Form } from "react-bootstrap";

import BaseComponent from "../components/Common/BaseComponent";
import React from "react";
import RendererFrame from "./Common/RendererFrame";

export default class Renderer extends BaseComponent {
  constructor(props) {
    super(props);
    this.state = {
      reloadOnRender: true,
    };
  }

  componentDidMount = async () => {
    await this.reloadWebAssembly();
  };

  reloadWebAssembly = async () => {
    window.render = undefined;

    let result = await WebAssembly.instantiateStreaming(
      fetch("http://localhost:8090/main.wasm"),
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

  onRenderClicked = async () => {
    if (this.state.reloadOnRender) {
      await this.reloadWebAssembly();
    }

    let output = window.render("param1", "param2"); // Exposed from golang
    console.log(output);
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
          <RendererFrame width="500" height="500"></RendererFrame>
        </Row>
      </Container>
    );
  }
}
