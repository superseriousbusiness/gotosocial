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

/*
 TODO: refactor dev-server to act as developer-facing webserver,
 proxying other requests to testrig instance. That way actual livereload works
*/

const Promise = require("bluebird");
const path = require('path');
const browserify = require("browserify");
const babelify = require('babelify');
const fsSync = require("fs");
const fs = require("fs").promises;
const chalk = require("chalk");

const devMode = process.env.NODE_ENV == "development";
if (devMode) {
	console.log(chalk.yellow("GoToSocial web asset bundler, running in development mode"));
} else {
	console.log(chalk.yellow("GoToSocial web asset bundler, creating production build"));
}

function out(name = "") {
	return path.join(__dirname, "../assets/dist/", name);
}

if (!fsSync.existsSync(out())){
	fsSync.mkdirSync(out(), { recursive: true });
}

module.exports = {out};

const splitCSS = require("./lib/split-css.js");

let cssFiles = fsSync.readdirSync(path.join(__dirname, "./css")).map((file) => {
	return path.join(__dirname, "./css", file);
});

const bundles = [
	{
		outputFile: "frontend.js",
		entryFiles: ["./frontend/index.js"],
		babelOptions: {
			global: true,
			exclude: /node_modules\/(?!photoswipe-dynamic-caption-plugin)/,
		}
	},
	{
		outputFile: "react-bundle.js",
		factors: {
			"./settings/index.js": "settings.js",
		}
	},
	{
		outputFile: "_delete", // not needed, we only care for the css that's already split-out by css-extract
		entryFiles: cssFiles,
	}
];

const postcssPlugins = [
	"postcss-import",
	"postcss-nested",
	"autoprefixer",
	"postcss-custom-prop-vars",
	"postcss-color-mod-function"
].map((plugin) => require(plugin)());

function browserifyConfig({transforms = [], plugins = [], babelOptions = {}}) {
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
		fullPaths: devMode,
		debug: devMode
	};
}

Promise.each(bundles, (bundleCfg) => {
	let transforms, plugins, entryFiles;
	let { outputFile, babelOptions} = bundleCfg;

	if (bundleCfg.factors != undefined) {
		let factorBundle = [require("factor-bundle"), {
			outputs: Object.values(bundleCfg.factors).map((file) => {
				return out(file);
			})
		}];

		plugins = [factorBundle];

		entryFiles = Object.keys(bundleCfg.factors);
	} else {
		entryFiles = bundleCfg.entryFiles;
	}

	let config = browserifyConfig({transforms, plugins, babelOptions, entryFiles, outputFile});

	return Promise.try(() => {
		return browserify(entryFiles, config);
	}).then((bundler) => {
		Promise.promisifyAll(bundler);

		function makeBundle(cause) {
			if (cause != undefined) {
				console.log(chalk.yellow(`Watcher: update on ${cause}, re-bundling`));
			}
			return Promise.try(() => {
				return bundler.bundleAsync();
			}).then((bundle) => {
				if (outputFile != "_delete") {
					console.log(chalk.magenta(`JS: writing to assets/dist/${outputFile}`));
					if (bundleCfg.factors != undefined) {
						Object.values(bundleCfg.factors).forEach((factor) => {
							console.log(chalk.magenta(`JS: writing to assets/dist/${factor}`));
						});
					}
					return fs.writeFile(out(outputFile), bundle);
				}
			}).catch((e) => {
				console.log(chalk.red("Fatal error in bundler:"), bundleCfg.outputFile);
				if (e.name == "CssSyntaxError") {
					// contains useful info about error + location, but followed by useless
					// actual stacktrace, so cut that off
					let stack = e.stack;
					stack.split("\n").some((line) => {
						if (line.startsWith("    at Input.error")) {
							return true;
						} else {
							console.log(line);
							return false;
						}
					});
				} else {
					console.log(e.message);
				}
				console.log();
			});
		}

		if (devMode) {
			bundler.on("update", makeBundle);
		}
		return makeBundle();
	});
}).then(() => {
	if (devMode) {
		console.log(chalk.yellow("Initial build finished, waiting for file changes"));
	} else {
		console.log(chalk.yellow("Finished building"));
	}
});

