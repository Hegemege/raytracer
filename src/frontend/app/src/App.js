import "./App.scss";

import { Col, Container, Nav, Navbar } from "react-bootstrap";
import {
  Link,
  Redirect,
  Route,
  BrowserRouter as Router,
  Switch,
} from "react-router-dom";
import { faImage, faCopy } from "@fortawesome/free-solid-svg-icons";

import BaseComponent from "./components/Common/BaseComponent";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import React from "react";
import { library } from "@fortawesome/fontawesome-svg-core";
import Renderer from "./components/Renderer";

library.add(faImage, faCopy);

export default class App extends BaseComponent {
  constructor(props) {
    super(props);
    this.state = {};
  }

  componentDidMount = async () => {};

  render() {
    return (
      <Container fluid>
        <Col>
          <Router>
            <Navbar variant="dark" expand="lg">
              <Nav className="mr-auto">
                <Nav.Link as={Link} to="/renderer">
                  <FontAwesomeIcon icon={["fas", "image"]} /> Renderer
                </Nav.Link>
              </Nav>
            </Navbar>

            <Switch>
              <Route exact path="/">
                <Redirect to="/renderer" />
              </Route>
              <Route exact path="/renderer">
                <Renderer></Renderer>
              </Route>
            </Switch>
          </Router>
        </Col>
      </Container>
    );
  }
}
