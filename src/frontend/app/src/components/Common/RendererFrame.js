import BaseComponent from "./BaseComponent";
import React from "react";

export default class RendererFrame extends BaseComponent {
  constructor(props) {
    super(props);
    this.state = {
      dirty: false,
    };
    this.canvasRef = React.createRef();
    this.imageData = null;
  }

  componentDidMount = async () => {
    let context = this.canvasRef.current.getContext("2d");
    this.imageData = context.createImageData(
      this.props.width,
      this.props.height
    );
    await this.setStateAsync({ ...this.state, dirty: true });
  };

  componentDidUpdate = async () => {
    if (this.state.dirty) {
      let context = this.canvasRef.current.getContext("2d");
      context.putImageData(this.imageData, 0, 0);
      await this.setStateAsync({ ...this.state, dirty: false });
    }
  };

  render() {
    return (
      <div className="renderer-frame">
        <canvas
          ref={this.canvasRef}
          width={this.props.width}
          height={this.props.height}
        ></canvas>
      </div>
    );
  }
}
