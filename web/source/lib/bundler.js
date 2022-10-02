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

const Promise = require("bluebird");
const browserify = require("browserify");
const babelify = require('babelify');
const chalk = require("chalk");
const fs = require("fs").promises;
const { EventEmitter } = require("events");
const path = require("path");
const debugLib = require("debug");
debugLib.enable("GoToSocial");
const debug = debugLib("GoToSocial");

const outputEmitter = new EventEmitter();

const splitCSS = require("./split-css")(outputEmitter);
const out = require("./output-path");

const postcssPlugins = [
	"postcss-import",
	"postcss-nested",
	"autoprefixer",
	"postcss-custom-prop-vars",
	"postcss-color-mod-function"
].map((plugin) => require(plugin)());

function browserifyConfig(devMode, { transforms = [], plugins = [], babelOptions = {} }) {
	if (devMode) {
		plugins.push(require("watchify"));
	} else {
		transforms.push([
			require("uglifyify"), {
				global: true,
				exts: ".js"
			}
		]);
	}

	return {
		cache: {},
		packageCache: {},
		transform: [
			[
				babelify.configure({
					presets: [
						[
							require.resolve("@babel/preset-env"),
							{
								modules: "cjs"
							}
						],
						require.resolve("@babel/preset-react")
					]
				}),
				babelOptions
			],
			...transforms
		],
		plugin: [
			[require("icssify"), {
				parser: require("postcss-scss"),
				before: postcssPlugins,
				mode: 'global'
			}],
			[require("css-extract"), { out: splitCSS }],
			...plugins
		],
		extensions: [".js", ".jsx", ".css"],
		basedir: path.join(__dirname, "../"),
		fullPaths: devMode,
		debug: devMode
	};
}

module.exports = function gtsBundler(devMode, bundles) {
	if (devMode) {
		require("./dev-server")(outputEmitter);
	}

	Promise.each(bundles, (bundleCfg) => {
		let transforms, plugins, entryFiles;
		let { outputFile, babelOptions } = bundleCfg;

		if (bundleCfg.factors != undefined) {
			let factorBundle = [require("factor-bundle"), {
				outputs: Object.values(bundleCfg.factors).map((file) => {
					return out(file);
				}),
				threshold: function(row, groups) {
					// always put livereload.js in common bundle
					if (row.id.endsWith("web/source/lib/livereload.js")) {
						return true;
					} else {
						return this._defaultThreshold(row, groups);
					}
				}
			}];

			plugins = [factorBundle];

			entryFiles = Object.keys(bundleCfg.factors);
		} else {
			entryFiles = bundleCfg.entryFiles;
		}

		if (devMode) {
			entryFiles.push(path.join(__dirname, "./livereload.js"));
		}

		let config = browserifyConfig(devMode, { transforms, plugins, babelOptions, entryFiles, outputFile });

		return Promise.try(() => {
			return browserify(entryFiles, config);
		}).then((bundler) => {
			Promise.promisifyAll(bundler);

			function makeBundle(cause) {
				if (cause != undefined) {
					debug(chalk.yellow(`Watcher: update on ${cause}, re-bundling`));
				}
				return Promise.try(() => {
					return bundler.bundleAsync();
				}).then((bundle) => {
					if (outputFile != "_delete") {
						let updates = new Set([outputFile]);
						if (bundleCfg.factors != undefined) {
							Object.values(bundleCfg.factors).forEach((factor) => {
								updates.add(factor);
								debug(chalk.magenta(`JS: writing to assets/dist/${factor}`));
							});
						}
						outputEmitter.emit("update", {type: "JS", updates: Array.from(updates)});
						return fs.writeFile(out(outputFile), bundle);
					}
				}).catch((e) => {
					debug(chalk.red("Fatal error in bundler:"), bundleCfg.outputFile);
					if (e.name == "CssSyntaxError") {
						// contains useful info about error + location, but followed by useless
						// actual stacktrace, so cut that off
						let stack = e.stack;
						stack.split("\n").some((line) => {
							if (line.startsWith("    at Input.error")) {
								return true;
							} else {
								debug(line);
								return false;
							}
						});
					} else {
						debug(e.message);
					}
					debug();
				});
			}

			if (devMode) {
				bundler.on("update", makeBundle);
			}
			return makeBundle();
		});
	}).then(() => {
		if (devMode) {
			debug(chalk.yellow("Initial build finished, waiting for file changes"));
		} else {
			debug(chalk.yellow("Finished building"));
		}
	});
};

outputEmitter.on("update", (u) => {
	u.updates.forEach((outputFile) => {
		let color = (str) => str;
		if (u.type == "JS") {
			color = chalk.magenta;
		} else {
			color = chalk.blue;
		}
		debug(color(`${u.type}: writing to assets/dist/${outputFile}`));
	});
});