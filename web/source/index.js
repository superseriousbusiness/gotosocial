"use strict";

const skulk = require("skulk");
const fs = require("fs");
const path = require("path");

let cssEntryFiles = fs.readdirSync(path.join(__dirname, "./css")).map((file) => {
	return path.join(__dirname, "./css", file);
});

const prodCfg = {
	transform: [
		["uglifyify", {
			global: true,
			exts: ".js"
		}],
		["@browserify/envify", {global: true}]
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
			proxy: "http://localhost:8081",
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
		settings: {
			entryFile: "settings",
			outputFile: "settings.js",
			prodCfg: prodCfg,
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