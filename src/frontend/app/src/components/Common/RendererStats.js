import BaseComponent from "./BaseComponent";
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
      total += this.props.data[key].timer;
    }
    return (
      <div className="render-stats">
        <Row className="render-stat-header-row">
          <Col md={2}>Progress</Col>
          <Col md={3}>Time</Col>
          <Col>Event</Col>
        </Row>
        {Object.keys(this.props.data ? this.props.data : {}).map((key, i) => (
          <Row key={i} className="render-stat-row">
            <div
              style={{
                position: "absolute",
                height: "2px",
                backgroundColor: "lightblue",
                width: (this.props.data[key].progress * 100).toString() + "%",
              }}
            ></div>
            <Col md={2}>{parseInt(this.props.data[key].progress * 100)}%</Col>
            <Col md={3}>{this.props.data[key].timer} ms</Col>
            <Col>{key}</Col>
          </Row>
        ))}
        <Row className="render-stat-total-row">
          <Col md={2}>Total</Col>
          <Col md={3}>{total} ms</Col>
          <Col></Col>
        </Row>
      </div>
    );
  }
}
