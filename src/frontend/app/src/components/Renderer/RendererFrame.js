import BaseComponent from "../Common/BaseComponent";
import React from "react";

export default class RendererFrame extends BaseComponent {
  constructor(props) {
    super(props);
    this.state = {
      renderedBlocks: 0,
      renderKey: null,
    };
    this.canvasRef = React.createRef();
    this.imageData = null;
    this.offscreenCanvas = null;
  }

  componentDidMount = async () => {};

  componentDidUpdate = async () => {
    if (!this.props.imageData) {
      return;
    }

    if (this.props.imageData.length == this.state.renderedBlocks) {
      return;
    }

    if (
      this.props.imageDetails &&
      this.props.renderKey !== this.state.renderKey
    ) {
      await this.setStateAsync({
        ...this.state,
        renderKey: this.props.renderKey,
      });
      this.initializeCanvases();
    }

    await this.setStateAsync({
      ...this.state,
      renderedBlocks: this.props.imageData.length,
    });

    this.updateCanvas();
  };

  initializeCanvases = () => {
    // Load the image data and put it into the canvas
    let width = this.props.imageDetails.width;
    let height = this.props.imageDetails.height;
    let scale = this.props.imageDetails.scale / 100.0;

    let offscreen = new OffscreenCanvas(width, height);
    this.offscreenCanvas = offscreen;

    let offscreenContext = this.offscreenCanvas.getContext("2d");
    offscreenContext.putImageData(
      new ImageData(
        new Uint8ClampedArray(4 * width * height).fill(0),
        width,
        height
      ),
      0,
      0
    );

    let context = this.canvasRef.current.getContext("2d");
    context.canvas.width = width * scale;
    context.canvas.height = height * scale;

    context.drawImage(
      this.offscreenCanvas,
      0,
      0,
      context.canvas.width,
      context.canvas.height
    );
  };

  updateCanvas = () => {
    for (let result of this.props.imageData) {
      let width = result.params.Width;
      let height = result.params.Height;
      let xoffset = result.params.XOffset;
      let yoffset = result.params.YOffset;

      let offscreenContext = this.offscreenCanvas.getContext("2d");

      let data = Uint8ClampedArray.from(
        this.base64ToArrayBuffer(result.imageData.Pix)
      );
      this.imageData = new ImageData(data, width, height);

      offscreenContext.putImageData(this.imageData, xoffset, yoffset);

      let context = this.canvasRef.current.getContext("2d");

      context.drawImage(
        this.offscreenCanvas,
        0,
        0,
        context.canvas.width,
        context.canvas.height
      );
    }
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
      <div>
        <canvas className="renderer-frame" ref={this.canvasRef} />
      </div>
    );
  }
}
