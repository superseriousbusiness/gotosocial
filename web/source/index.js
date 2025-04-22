/*
	GoToSocial
	Copyright (C) GoToSocial Authors admin@gotosocial.org
	SPDX-License-Identifier: AGPL-3.0-or-later

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

const skulk = require("skulk");
const fs = require("fs");
const path = require("path");

let cssEntryFiles = fs.readdirSync(path.join(__dirname, "./css")).map((file) => {
	return path.join(__dirname, "./css", file);
});

const prodCfg = {
	transform: [
		["@browserify/uglifyify", {
			global: true,
			exts: ".js"
		}],
		["@browserify/envify", { global: true }]
	]
};

skulk({
	name: "GoToSocial",
	basePath: __dirname,
	assetPath: "../assets/",
	prodCfg: {
		servers: {
			express: false,
			livereload: false
		}
	},
	servers: {
		express: {
			proxy: "http://127.0.0.1:8081",
			assets: "/assets"
		}
	},
	bundles: {
		frontend: {
			entryFile: "frontend",
			outputFile: "frontend.js",
			preset: ["js"],
			prodCfg: prodCfg,
			transform: [
				["babelify", {
					global: true,
					ignore: [/node_modules\/(?!(photoswipe.*))/]
				}]
			],
		},
		frontend_prerender: {
			entryFile: "frontend_prerender",
			outputFile: "frontend_prerender.js",
			preset: ["js"],
			prodCfg: prodCfg,
			transform: [
				["babelify", { global: true }]
			],
		},
		settings: {
			entryFile: "settings",
			outputFile: "settings.js",
			prodCfg: prodCfg,
			plugin: [
				// Additional settings for TS are passed from tsconfig.json.
				// See: https://github.com/TypeStrong/tsify#tsconfigjson
				["tsify"]
			],
			transform: [
				// tsify is called before babelify, so we're just babelifying
				// commonjs here, no need for the typescript preset.
				["babelify", {
					global: true,
					ignore: [/node_modules\/(?!(nanoid)|(wouter))/],
				}]
			],
			presets: [
				"react",
				["postcss", {
					output: "settings-style.css"
				}]
			]
		},
		css: {
			entryFiles: cssEntryFiles,
			outputFile: "_discard",
			presets: [["postcss", {
				output: "_split"
			}]]
		}
	}
});
