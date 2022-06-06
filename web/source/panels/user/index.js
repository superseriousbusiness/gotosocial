"use strict";

const Promise = require("bluebird");
const React = require("react");
const ReactDom = require("react-dom");

// require("./style.css");

function App() {
	return "hello world - user panel";
}

ReactDom.render(<App/>, document.getElementById("root"));