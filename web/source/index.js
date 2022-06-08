/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

"use strict";

/*
	Bundle the frontend panels for admin and user settings
*/

const path = require('path');
const budoExpress = require('@f0x52/budo-express');
const babelify = require('babelify');
const icssify = require("icssify");
const fs = require("fs");
const EventEmitter = require('events');

function out(name = "") {
	return path.join(__dirname, "../assets/dist/", name);
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
	transform: babelify.configure({ presets: [require.resolve("@babel/preset-env"), require.resolve("@babel/preset-react")] }),
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

const server = budoExpress({
	port: 8081,
	host: "localhost",
	entryFiles: entryFiles,
	basePath: __dirname,
	bundlePath: "bundle.js",
	staticPath: out(),
	expressApp: require("./dev-server.js"),
	browserify: browserifyConfig,
	livereloadPattern: "**/*.{html,js,svg}"
});

if (server instanceof EventEmitter) {
	server.on("update", (contents) => {
	console.log("writing bundle.js to dist/");
	fs.writeFileSync(out("bundle.js"), contents);
});
}