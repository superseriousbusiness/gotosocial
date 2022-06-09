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

const fs = require("fs");
const path = require("path");

const {Writable} = require("stream");
const {out} = require("../index.js");

const fromRegex = /\/\* from (.+?) \*\//;
module.exports = function splitCSS() {
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

			const cssDir = path.join(__dirname, "../css");

			stream.split("\n").forEach((line) => {
				if (line.startsWith("/* from")) {
					let found = fromRegex.exec(line);
					if (found != null) {
						write();

						let parts = path.parse(found[1]);
						if (path.relative(cssDir, path.join(process.cwd(), parts.dir)) == "") {
							input = parts.base;
						} else {
							// prefix filename with path
							let relative = path.relative(path.join(__dirname, "../"), path.join(process.cwd(), found[1]));
							input = relative.replace(/\//g, "-");
						}
					}
				} else {
					content.push(line);
				}
			});
			write();
		}
	});
};
