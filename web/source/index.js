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

module.exports = {out};

const splitCSS = require("./lib/split-css.js");

const bundles = {
	"./frontend/index.js": "frontend.js",
	"./panels/admin/index.js": "admin-panel.js",
	"./panels/user/index.js": "user-panel.js",
};

const postcssPlugins = [
	"postcss-import",
	"postcss-strip-inline-comments",
	"postcss-nested",
	"autoprefixer",
	"postcss-custom-prop-vars",
	"postcss-color-mod-function"
].map((plugin) => require(plugin)());

const browserifyConfig = {
	transform: [
		babelify.configure({
			presets: [
				require.resolve("@babel/preset-env"),
				require.resolve("@babel/preset-react")
			]
		}),
		["uglifyify", {
			global: true,
			exts: ".js"
		}]
	],
	plugin: [
		[icssify, {
			parser: require('postcss-scss'),
			before: postcssPlugins,
			mode: 'global'
		}],
		[require("css-extract"), { out: splitCSS }],
		[require("factor-bundle"), {
			outputs: Object.values(bundles).map((file) => {
				return out(file);
			})
		}]
	]
};

const entryFiles = Object.keys(bundles);

fs.readdirSync(path.join(__dirname, "./css")).forEach((file) => {
	entryFiles.push(path.join(__dirname, "./css", file));
});

if (!fs.existsSync(out())){
	fs.mkdirSync(out(), { recursive: true });
}

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
		fs.writeFileSync(out("bundle.js"), contents);
	});
}