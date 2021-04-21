import React from "react";

export default class BaseComponent extends React.Component {
  constructor(props) {
    super(props);
    this.state = {};
  }

  setStateAsync = (state) => {
    return new Promise((resolve) => {
      this.setState(state, resolve);
    });
  };
}
