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
	Bundle the PostCSS stylesheets and javascript bundles for general frontend and settings panel
*/

const path = require('path');
const fsSync = require("fs");
const chalk = require("chalk");

const gtsBundler = require("./lib/bundler");

const devMode = process.env.NODE_ENV == "development";
if (devMode) {
	console.log(chalk.yellow("GoToSocial web asset bundler, running in development mode"));
} else {
	console.log(chalk.yellow("GoToSocial web asset bundler, creating production build"));
}

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

return gtsBundler(devMode, bundles);