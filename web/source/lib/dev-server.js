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

const tinylr = require("tiny-lr");
const chalk = require("chalk");

const PORT = 35729;

module.exports = function devServer(outputEmitter) {
	let server = tinylr();
	
	server.listen(PORT, () => {
		console.log(chalk.cyan(`Livereload server listening on :${PORT}`));
	});

	outputEmitter.on("update", ({updates}) => {
		let fullPaths = updates.map((path) => `/assets/dist/${path}`);
		tinylr.changed(fullPaths.join(","));
	});

	process.on("SIGUSR2", server.close);
	process.on("SIGTERM", server.close);
};