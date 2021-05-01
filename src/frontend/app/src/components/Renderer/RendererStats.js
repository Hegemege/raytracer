import BaseComponent from "../Common/BaseComponent";
import React from "react";
import { Row, Col } from "react-bootstrap";

export default class RendererStats extends BaseComponent {
  constructor(props) {
    super(props);
    this.state = {};
  }

  componentDidMount = async () => {};

  componentDidUpdate = async () => {
    if (!this.props.data) {
      return;
    }
  };

  render() {
    let total = 0;
    for (let key in this.props.data) {
      for (let taskId in this.props.data[key]) {
        total += this.props.data[key][taskId].timer;
      }
    }

    let data = this.props.data;
    if (!data) {
      return null;
    }

    return (
      <div className="render-stats">
        <Row>Worker ID {this.props.workerId}</Row>
        <Row className="render-stat-header-row">
          <Col md={2}>Progress</Col>
          <Col md={3}>Time</Col>
          <Col>Event</Col>
        </Row>
        {Object.keys(data).map((key, i) =>
          Object.keys(data[key])
            .sort()
            .reverse()
            .map((taskId, j) => (
              <Row key={`${i}-${j}`} className="render-stat-row">
                <div
                  style={{
                    position: "absolute",
                    height: "2px",
                    backgroundColor: "lightblue",
                    width: (data[key][taskId].progress * 100).toString() + "%",
                  }}
                ></div>
                <Col md={2}>{parseInt(data[key][taskId].progress * 100)}%</Col>
                <Col md={3}>{data[key][taskId].timer} ms</Col>
                <Col>{key}</Col>
              </Row>
            ))
        )}
        <Row className="render-stat-total-row">
          <Col md={2}>Total</Col>
          <Col md={3}>{total} ms</Col>
          <Col></Col>
        </Row>
      </div>
    );
  }
}
