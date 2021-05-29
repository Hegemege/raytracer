import { Row, Button, Container, Form } from "react-bootstrap";

import BaseComponent from "../Common/BaseComponent";
import React from "react";

export default class RendererParams extends BaseComponent {
  constructor(props) {
    super(props);
    this.state = {
      presetName: undefined,
      params: {
        width: 500,
        height: 500,
        scale: 100,
        x: 0,
        y: 0,
        z: 0,
        rx: 0,
        ry: 0,
        rz: 0,
        projection: 0,
        projectionPlaneDistance: 1,
        fieldOfView: 60,
        ortographicSize: 10,
        bounces: 4,
        lightSampleRays: 8,
        raysPerPixel: 5,
        workerCount: 16,
        taskCount: 16,
        rngSeed: this.getNewSeed(),
        gammaCorrection: true,
        gamma: "2.2",
        useBVH: true,
        saveBVH: true,
        loadBVH: true,
        maxLeafSize: 6,
        maxDepth: 20,
        lightIntensity: 100,
        renderAfterInitialization: true,
        incrementalRendering: false,
        forceDebugLight: false,
        debugLightAtCamera: true,
        debugLightSize: 1.0,
        debugLightX: 0,
        debugLightY: 0,
        debugLightZ: 0,
        debugLightRX: 0,
        debugLightRY: 0,
        debugLightRZ: 0,
        debugLightCR: 255,
        debugLightCG: 255,
        debugLightCB: 255,
        // Note: when adding properties, add them to the preset.json files too
      },
    };

    this.projectionMap = ["Perspective", "Ortographic"];
  }

  componentDidMount = async () => {
    this.onParamsChanged();
  };

  componentDidUpdate = async () => {
    if (!this.props.params) {
      return;
    }

    let samePreset = this.state.paramKey === this.props.paramKey;

    if (samePreset) {
      return;
    }

    // Different preset or force rollback changes
    await this.setStateAsync({
      ...this.state,
      paramKey: this.props.paramKey,
      params: {
        ...this.state.params,
        ...this.props.params,
        rngSeed: this.props.params.rngSeed
          ? this.props.params.rngSeed
          : this.getNewSeed(),
      },
    });

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

  getIntParam = (field) => {
    let param = this.state.params[field];
    if (!param) {
      return 0; // Default value
    }

    return param;
  };

  getFloatParam = (field) => {
    let param = this.state.params[field];
    if (!param) {
      return 0; // Default value
    }

    return param;
  };

  getBoolParam = (field) => {
    let param = this.state.params[field];
    if (!param) {
      return false; // Default value
    }

    return param;
  };

  renderParam(field, type, label, size, disabled) {
    let controlId = `form${field}`;

    if (type === "int") {
      let value = this.getIntParam(field);
      return (
        <Form.Group controlId={controlId} className="right-margin">
          <Form.Label>{label}</Form.Label>
          <Form.Control
            disabled={disabled}
            htmlSize={size}
            type="text"
            label={label}
            value={value}
            onChange={(e) => this.handleIntParamChanged(e, field)}
          />
        </Form.Group>
      );
    } else if (type === "float") {
      let value = this.getFloatParam(field);
      return (
        <Form.Group controlId={controlId} className="right-margin">
          <Form.Label>{label}</Form.Label>
          <Form.Control
            disabled={disabled}
            htmlSize={size}
            type="text"
            label={label}
            value={value}
            onChange={(e) => this.handleFloatParamChanged(e, field)}
          />
        </Form.Group>
      );
    } else if (type === "bool") {
      let value = this.getBoolParam(field);
      return (
        <Form.Group controlId={controlId} className="right-margin">
          <Form.Check
            disabled={disabled}
            type="checkbox"
            label={label}
            checked={value}
            onChange={(e) => this.handleBoolParamChanged(e, field)}
          />
        </Form.Group>
      );
    } else {
      return null;
    }
  }

  render() {
    if (!this.state.params) {
      return null;
    }

    return (
      <Container>
        <Form>
          <Row className="param-row">
            {this.renderParam("width", "int", "Width px", 4)}
            {this.renderParam("height", "int", "Height px", 4)}
            {this.renderParam("scale", "int", "Scale %", 4)}

            <Form.Group controlId="formProjection" className="right-margin">
              <Form.Label>Projection</Form.Label>
              <Form.Control
                as="select"
                value={this.projectionMap[this.getIntParam("projection")]}
                onChange={this.handleProjectionChanged}
              >
                {this.projectionMap.map((item, index) => (
                  <option key={index}>{item}</option>
                ))}
              </Form.Control>
            </Form.Group>

            {this.renderParam(
              "projectionPlaneDistance",
              "float",
              "Proj. Distance",
              4
            )}

            {this.getIntParam("projection") === 0 &&
              this.renderParam("fieldOfView", "int", "Field of View", 6)}
            {this.getIntParam("projection") === 1 &&
              this.renderParam("ortographicSize", "int", "Ortographic Size", 6)}

            {this.renderParam("raysPerPixel", "int", "Rays Per Pixel", 6)}
            {this.renderParam("bounces", "int", "Bounces", 4)}
            {this.renderParam("lightSampleRays", "int", "Light Sample Rays", 6)}
          </Row>

          <Row className="param-row">
            <Form.Label className="header">Camera</Form.Label>
            {this.renderParam("x", "float", "X", 4)}
            {this.renderParam("y", "float", "Y", 4)}
            {this.renderParam("z", "float", "Z", 4)}
            {this.renderParam("rx", "float", "rX", 4)}
            {this.renderParam("ry", "float", "rY", 4)}
            {this.renderParam("rz", "float", "rZ", 4)}

            {this.renderParam("lightIntensity", "float", "Light Intensity", 6)}

            <Form.Group
              controlId="formCheckboxDebugLightToggle"
              className="right-margin"
            >
              <Form.Label>Debug Light</Form.Label>
              <Form.Check
                style={{ marginTop: "-10px" }}
                id="formCheckboxForceDebugLight"
                type="checkbox"
                label="Force Debug Light"
                checked={this.getBoolParam("forceDebugLight")}
                onChange={(e) =>
                  this.handleBoolParamChanged(e, "forceDebugLight")
                }
              />
              <Form.Check
                disabled={!this.getBoolParam("forceDebugLight")}
                id="formCheckboxDebugLightAtCamera"
                type="checkbox"
                label="Debug Light At Camera"
                checked={this.getBoolParam("debugLightAtCamera")}
                onChange={(e) =>
                  this.handleBoolParamChanged(e, "debugLightAtCamera")
                }
              />
            </Form.Group>
          </Row>
          <Row
            className="param-row"
            hidden={
              this.getBoolParam("debugLightAtCamera") ||
              !this.getBoolParam("forceDebugLight")
            }
          >
            <Form.Label className="header">Debug Light</Form.Label>
            {this.renderParam("debugLightX", "float", "X", 4)}
            {this.renderParam("debugLightY", "float", "Y", 4)}
            {this.renderParam("debugLightZ", "float", "Z", 4)}
            {this.renderParam("debugLightRX", "float", "rX", 4)}
            {this.renderParam("debugLightRY", "float", "rY", 4)}
            {this.renderParam("debugLightRZ", "float", "rZ", 4)}
            {this.renderParam("debugLightSize", "float", "Debug Light Size", 6)}
          </Row>
          <Row className="param-row bottom-align">
            {this.renderParam("workerCount", "int", "Workers", 6)}
            {this.renderParam(
              "taskCount",
              "int",
              "Tasks (n^2)",
              6,
              this.state.params.incrementalRendering === true
            )}

            {this.renderParam("rngSeed", "int", "RNG Seed", 22)}

            <Form.Group controlId="formGamma" className="right-margin">
              <Form.Check
                id="formCheckboxGammaCorrection"
                type="checkbox"
                label="Gamma correction"
                checked={this.getBoolParam("gammaCorrection")}
                onChange={(e) =>
                  this.handleBoolParamChanged(e, "gammaCorrection")
                }
              />
              <Form.Label>Gamma</Form.Label>
              <Form.Control
                htmlSize="6"
                type="text"
                label="Gamma"
                value={this.getFloatParam("gamma")}
                onChange={(e) => this.handleFloatParamChanged(e, "gamma")}
              />
            </Form.Group>
            <Form.Group controlId="formBVHSettings" className="right-margin">
              <Form.Check
                id="formCheckboxUseBVH"
                type="checkbox"
                label="Use BVH"
                checked={this.getBoolParam("useBVH")}
                onChange={(e) => this.handleBoolParamChanged(e, "useBVH")}
              />
              <Form.Label>Max Leaf Tris</Form.Label>
              <Form.Control
                htmlSize="3"
                type="text"
                label="Max Leaf Tris"
                value={this.getIntParam("maxLeafSize")}
                onChange={(e) => this.handleIntParamChanged(e, "maxLeafSize")}
              />
            </Form.Group>
            <Form.Group controlId="formBVHDepth" className="right-margin">
              <Form.Label>Max Depth</Form.Label>
              <Form.Control
                htmlSize="3"
                type="text"
                label="Max Depth"
                value={this.getIntParam("maxDepth")}
                onChange={(e) => this.handleIntParamChanged(e, "maxDepth")}
              />
            </Form.Group>
            <Form.Group
              controlId="formCheckboxSaveBVH"
              className="right-margin"
            >
              <Form.Label>Caching</Form.Label>
              <Form.Check
                id="formCheckboxSaveBVH"
                type="checkbox"
                label="Save BVH"
                checked={this.getBoolParam("saveBVH")}
                onChange={(e) => this.handleBoolParamChanged(e, "saveBVH")}
              />
              <Form.Check
                id="formCheckboxLoadBVH"
                type="checkbox"
                label="Load BVH"
                checked={this.getBoolParam("loadBVH")}
                onChange={(e) => this.handleBoolParamChanged(e, "loadBVH")}
              />
            </Form.Group>
          </Row>
          <Row className="param-row bottom-align">
            {this.renderParam(
              "renderAfterInitialization",
              "bool",
              "Render after initialization"
            )}
            {this.renderParam(
              "incrementalRendering",
              "bool",
              "Incremental rendering"
            )}
          </Row>
          <Row>
            <Button
              variant="primary"
              type="submit"
              className="right-margin"
              onClick={this.onInitializeClicked}
            >
              Initialize
            </Button>
            <Button
              variant="primary"
              className="right-margin"
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
