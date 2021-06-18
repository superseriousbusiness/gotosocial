"use strict";

const fs = require("fs").promises;
const postcss = require('postcss');
const {parse} = require("postcss-scss");

const postcssPlugins = ["postcss-strip-inline-comments", "postcss-nested", "postcss-simple-vars", "postcss-color-function"].map((plugin) => require(plugin)());

let inputFile = `${__dirname}/style.css`;
let outputFile = `${__dirname}/../assets/bundle.css`;

fs.readFile(inputFile, "utf-8").then((input) => {
	return parse(input);
}).then((ast) => {
	return postcss(postcssPlugins).process(ast, {
		from: "style.css",
		to: "bundle.css"
	});
}).then((bundle) => {
	return fs.writeFile(outputFile, bundle.css);
}).then(() => {
	console.log("Finished writing CSS bundle");
});
