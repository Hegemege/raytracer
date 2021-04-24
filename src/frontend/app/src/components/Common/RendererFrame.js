import BaseComponent from "./BaseComponent";
import React from "react";

export default class RendererFrame extends BaseComponent {
  constructor(props) {
    super(props);
    this.state = {
      renderedData: "",
    };
    this.canvasRef = React.createRef();
    this.imageData = null;
  }

  componentDidMount = async () => {};

  componentDidUpdate = async () => {
    if (!this.props.imageData) {
      return;
    }

    if (this.props.imageData.Pix == this.state.renderedData) {
      return;
    }

    await this.setStateAsync({
      ...this.state,
      renderedData: this.props.imageData.Pix,
    });

    this.updateCanvas();
  };

  updateCanvas = () => {
    // Load the image data and put it into the canvas
    let context = this.canvasRef.current.getContext("2d");

    let width = this.props.imageData.Rect.Max.X;
    let height = this.props.imageData.Rect.Max.Y;

    context.canvas.width = width;
    context.canvas.height = height;

    let data = Uint8ClampedArray.from(
      this.base64ToArrayBuffer(this.props.imageData.Pix)
    );
    this.imageData = new ImageData(data, width, height);

    context.putImageData(this.imageData, 0, 0);
  };

  base64ToHex = (str) => {
    const raw = atob(str);
    let result = "";
    for (let i = 0; i < raw.length; i++) {
      const hex = raw.charCodeAt(i).toString(16);
      result += hex.length === 2 ? hex : "0" + hex;
    }
    return result.toUpperCase();
  };

  base64ToArrayBuffer = (base64) => {
    var binary_string = atob(base64);
    var len = binary_string.length;
    var bytes = new Uint8ClampedArray(len);
    for (var i = 0; i < len; i++) {
      bytes[i] = binary_string.charCodeAt(i);
    }
    return bytes;
  };

  render() {
    return (
      <div className="renderer-frame">
        <canvas ref={this.canvasRef} />
      </div>
    );
  }
}
