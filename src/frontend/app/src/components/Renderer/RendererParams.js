import { Row, Button, Container, Form } from "react-bootstrap";

import BaseComponent from "../Common/BaseComponent";
import React from "react";

export default class RendererParams extends BaseComponent {
  constructor(props) {
    super(props);
    this.state = {
      params: {
        width: 100,
        height: 100,
        scale: 500,
        x: -0.225,
        y: 2.55,
        z: 6,
        rx: 0,
        ry: 0,
        rz: 0,
        projection: 0,
        fieldOfView: 45,
        ortographicSize: 3,
        bounces: 10,
        lightSampleRays: 4,
        raysPerPixel: 10,
        sceneData: null,
        objData: "",
        mtlData: "",
        workerCount: 16,
        taskCount: 16,
        rngSeed: this.getNewSeed(),
        gammaCorrection: true,
        gamma: "2.2",
        useBVH: true,
        maxLeafSize: 6,
        lightIntensity: 100,
        debugLightSize: 1.0,
        renderAfterInitialization: true,
        incrementalRendering: false,
      },
    };
  }

  componentDidMount = async () => {
    // TODO: Save/load config from localstorage/json
    this.onParamsChanged();
  };

  onParamsChanged = async () => {
    if (this.state.params.incrementalRendering) {
      await this.setStateAsync({
        ...this.state,
        params: {
          ...this.state.params,
          taskCount: this.state.params.workerCount,
        },
      });
    }

    await this.props.onChanged({ ...this.state.params });
  };

  handleBoolParamChanged = async (event, field) => {
    let value = event.target.checked;
    let params = {
      ...this.state.params,
    };
    params[field] = value;
    await this.setStateAsync({
      ...this.state,
      params: params,
    });

    await this.onParamsChanged();
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

    await this.onParamsChanged();
  };

  handleFloatParamChanged = async (event, param) => {
    const floatRegExp = new RegExp("^[+-]?([0-9]+([.][0-9]*)?|[.][0-9]+)$");
    if (event.target.value !== "" && !floatRegExp.test(event.target.value)) {
      return;
    }

    let params = {
      ...this.state.params,
    };
    params[param] = event.target.value;
    await this.setStateAsync({
      ...this.state,
      params: params,
    });

    await this.onParamsChanged();
  };

  handleProjectionChanged = async (event) => {
    await this.setStateAsync({
      ...this.state,
      params: {
        ...this.state.params,
        projection: event.target.options.selectedIndex,
      },
    });

    await this.onParamsChanged();
  };

  onAbortClicked = async (e) => {
    await this.props.onAbort(e);
  };

  onRenderClicked = async (e) => {
    await this.props.onStartRender(e, { ...this.state.params });
  };

  onInitializeClicked = async (e) => {
    e.preventDefault();

    if (e.detail === 0) {
      // Enter key pressed
      if (this.props.initialized) {
        await this.props.onStartRender(e, { ...this.state.params });
      } else {
        await this.props.onInitializeContext(e, { ...this.state.params });
      }
    } else {
      // Button clicked, force initialization
      await this.props.onInitializeContext(e, { ...this.state.params });
    }
  };

  getNewSeed = () => {
    return Math.floor(Math.random() * Number.MAX_SAFE_INTEGER);
  };

  render() {
    return (
      <Container>
        <Form>
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
                defaultValue={this.state.params.projection}
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

            <Form.Group controlId="formLightSampleRays" className="form-margin">
              <Form.Label>Light Sample Rays</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Light Sample Rays"
                value={this.state.params.lightSampleRays}
                onChange={(e) =>
                  this.handleIntParamChanged(e, "lightSampleRays")
                }
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
                value={this.state.params.x}
                onChange={(e) => this.handleFloatParamChanged(e, "x")}
              />
            </Form.Group>
            <Form.Group controlId="formY" className="form-margin">
              <Form.Label>Y</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                value={this.state.params.y}
                onChange={(e) => this.handleFloatParamChanged(e, "y")}
              />
            </Form.Group>
            <Form.Group controlId="formZ" className="form-margin">
              <Form.Label>Z</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                value={this.state.params.z}
                onChange={(e) => this.handleFloatParamChanged(e, "z")}
              />
            </Form.Group>

            <Form.Group controlId="formrX" className="form-margin">
              <Form.Label>rX</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                value={this.state.params.rx}
                onChange={(e) => this.handleFloatParamChanged(e, "rx")}
              />
            </Form.Group>
            <Form.Group controlId="formrY" className="form-margin">
              <Form.Label>rY</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                value={this.state.params.ry}
                onChange={(e) => this.handleFloatParamChanged(e, "ry")}
              />
            </Form.Group>
            <Form.Group controlId="formrZ" className="form-margin">
              <Form.Label>rZ</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Scale"
                value={this.state.params.rz}
                onChange={(e) => this.handleFloatParamChanged(e, "rz")}
              />
            </Form.Group>

            <Form.Group controlId="formLightIntensity" className="form-margin">
              <Form.Label>Light Intensity</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Light Intensity"
                value={this.state.params.lightIntensity}
                onChange={(e) =>
                  this.handleFloatParamChanged(e, "lightIntensity")
                }
              />
            </Form.Group>

            <Form.Group controlId="formDebugLightSize" className="form-margin">
              <Form.Label>Debug Light Size</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Debug Light Size"
                value={this.state.params.debugLightSize}
                onChange={(e) =>
                  this.handleFloatParamChanged(e, "debugLightSize")
                }
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
                Tasks (n<sup>2</sup>)
              </Form.Label>
              <Form.Control
                disabled={this.state.params.incrementalRendering}
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

            <Form.Group
              controlId="formCheckboxGammaCorrection"
              className="form-margin"
            >
              <Form.Check
                type="checkbox"
                label="Gamma correction"
                checked={this.state.params.gammaCorrection}
                onChange={(e) =>
                  this.handleBoolParamChanged(e, "gammaCorrection")
                }
              />
            </Form.Group>

            <Form.Group controlId="formGamma" className="form-margin">
              <Form.Label>Gamma</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Gamma"
                value={this.state.params.gamma}
                onChange={(e) => this.handleFloatParamChanged(e, "gamma")}
              />
            </Form.Group>
            <Form.Group controlId="formCheckboxUseBVH" className="form-margin">
              <Form.Check
                type="checkbox"
                label="Use BVH"
                checked={this.state.params.useBVH}
                onChange={(e) => this.handleBoolParamChanged(e, "useBVH")}
              />
            </Form.Group>
            <Form.Group controlId="formMaxLeafSize" className="form-margin">
              <Form.Label>Max Leaf Tris</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Max Leaf Tris"
                value={this.state.params.maxLeafSize}
                onChange={(e) => this.handleIntParamChanged(e, "maxLeafSize")}
              />
            </Form.Group>
          </Row>
          <Row>
            <Form.Group
              controlId="formCheckboxRenderAfterInitialization"
              className="form-margin"
            >
              <Form.Check
                type="checkbox"
                label="Render after initialization"
                checked={this.state.params.renderAfterInitialization}
                onChange={(e) =>
                  this.handleBoolParamChanged(e, "renderAfterInitialization")
                }
              />
            </Form.Group>
            <Form.Group
              controlId="formCheckboxIncrementalRendering"
              className="form-margin"
            >
              <Form.Check
                type="checkbox"
                label="Incremental rendering"
                checked={this.state.params.incrementalRendering}
                onChange={(e) =>
                  this.handleBoolParamChanged(e, "incrementalRendering")
                }
              />
            </Form.Group>
          </Row>
          <Row>
            <Button
              variant="primary"
              type="submit"
              className="form-margin"
              onClick={this.onInitializeClicked}
            >
              Initialize
            </Button>
            <Button
              variant="primary"
              className="form-margin"
              onClick={this.onRenderClicked}
              disabled={!this.props.initialized || this.props.aborted}
            >
              Render
            </Button>
            {this.props.running ? (
              <Button variant="danger" onClick={this.onAbortClicked}>
                Abort
              </Button>
            ) : null}
          </Row>
        </Form>
      </Container>
    );
  }
}
