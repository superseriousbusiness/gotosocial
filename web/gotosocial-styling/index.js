"use strict";

const Promise = require("bluebird");
const fs = require("fs").promises;
const postcss = require('postcss');
const {parse} = require("postcss-scss");
const argv = require('minimist')(process.argv.slice(2));

/*
	Bundle all postCSS files under the `templates/` directory separately, each prepended with the (variable) contents of ./colors.css
	Outputs in plain CSS are in `build/`, split by template
*/

const postcssPlugins = ["postcss-strip-inline-comments", "postcss-nested", "postcss-simple-vars", "postcss-color-function"].map((plugin) => require(plugin)());

function getTemplates() {
	return fs.readdir(`${__dirname}/templates`).then((templates) => {
		return templates.map((a) => {
			return [a, `${__dirname}/templates/${a}`];
		});
	});
}

getTemplates();

function bundle([template, path]) {
	return Promise.try(() => {
		return Promise.all([
			fs.readFile(`${__dirname}/colors.css`, "utf-8"),
			fs.readFile(path, "utf-8")
		]);
	}).then(([colors, style]) => {
		return parse(colors + "\n" + style);
	}).then((ast) => {
		return postcss(postcssPlugins).process(ast, {
			from: template,
			to: template
		});
	}).then((bundle) => {
		return fs.writeFile(`${buildDir}/${template}`, bundle.css);
	}).then(() => {
		console.log(`Finished writing CSS to ${buildDir}/${template}`);
	});
}

let buildDir

// try reading from arguments first
if (argv["build-dir"] != undefined) {
	buildDir = argv["build-dir"]
}

// then try reading from environment variable
if (buildDir == undefined) {
	buildDir = process.env.BUILD_DIR;
}

// then take default
if (buildDir == undefined) {
	buildDir = `${__dirname}/build`;
}

console.log("bundling to", buildDir);

function bundleAll() {
	return getTemplates().then((templates) => {
		return Promise.map(templates, bundle);
	});
}

if (process.env.NODE_ENV != "development") {
	bundleAll();
} else {
	const chokidar = require("chokidar");
	console.log("Watching for changes");
	chokidar.watch(`${__dirname}/templates`).on("all", (_, path) => {
		if (path.endsWith(".css")) {
			bundle([path.split("/").slice(-1)[0], path]);
		}
	});
	chokidar.watch(`${__dirname}/colors.css`).on("all", () => {
		console.log("colors.css updated, rebuilding all templates");
		bundleAll();
	});
}