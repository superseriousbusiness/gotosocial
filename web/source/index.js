"use strict";

/*
	Bundle the frontend panels for admin and user settings
*/

const path = require('path');
const budoExpress = require('budo-express');
const babelify = require('babelify');
const icssify = require("icssify");
const fs = require("fs");

const {Writable} = require("stream");

function out(name = "") {
	return path.join(__dirname, "../assets/bundled/", name);
}

const postcssPlugins = [
	"postcss-import",
	"postcss-strip-inline-comments",
	"postcss-nested",
	"autoprefixer",
	"postcss-custom-prop-vars",
	"postcss-color-mod-function"
].map((plugin) => require(plugin)());

const fromRegex = /\/\* from (.+?) \*\//;
function splitCSS() {
	let chunks = [];
	return new Writable({
		write: function(chunk, encoding, next) {
			chunks.push(chunk);
			next();
		},
		final: function() {
			let stream = chunks.join("");
			let input;
			let content = [];

			function write() {
				if (content.length != 0) {
					if (input == undefined) {
						throw new Error("Got CSS content without filename, can't output: ", content);
					} else {
						console.log("writing to", out(input));
						fs.writeFileSync(out(input), content.join("\n"));
					}
					content = [];
				}
			}

			stream.split("\n").forEach((line) => {
				if (line.startsWith("/* from")) {
					let found = fromRegex.exec(line);
					if (found != null) {
						write();

						let parts = path.parse(found[1]);
						if (parts.dir == "css") {
							input = parts.base;
						} else {
							input = found[1].replace(/\//g, "-");
						}
					}
				} else {
					content.push(line);
				}
			});
			write();
		}
	});
}

const browserifyConfig = {
	transform: babelify.configure({ presets: ["@babel/preset-env", "@babel/preset-react"] }),
	plugin: [
		[icssify, {
			parser: require('postcss-scss'),
			before: postcssPlugins,
			mode: 'global'
		}],
		[require("css-extract"), { out: splitCSS }],
		[require("factor-bundle"), { outputs: [out("/admin-panel.js"), out("user-panel.js")] }]
	]
};

const entryFiles = [
	'./panels/admin/index.js',
	'./panels/user/index.js',
];

fs.readdirSync(path.join(__dirname, "./css")).forEach((file) => {
	entryFiles.push(path.join(__dirname, "./css", file));
});

budoExpress({
	port: 8080,
	host: "10.0.1.1",
	allowUnsafeHost: true, // FIXME: don't do that by default lol
	entryFiles: entryFiles,
	basePath: __dirname,
	bundlePath: "bundle.js",
	staticPath: out(),
	expressApp: require("./dev-server.js"),
	browserify: browserifyConfig,
	livereloadPattern: "**/*.{html,js,svg}"
});