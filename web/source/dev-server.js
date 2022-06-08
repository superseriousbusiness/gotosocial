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

const express = require("express");

const app = express();

function html(title, css, js) {
	return `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
	${["_colors.css", "base.css", ...css].map((file) => {
		return `<link rel="stylesheet" href="/${file}"></link>`;
	}).join("\n")}
		<title>GoToSocial ${title} Panel</title>
	</head>
	<body>
		<header>
			<img src="/assets/logo.png" alt="Instance Logo">
			<div>
				<h1>
					GoToSocial ${title} Panel
				</h1>
			</div>
		</header>
		<main class="lightgray">
			<div id="root"></div>
		</main>
	${["bundle.js", ...js].map((file) => {
		return `<script src="/${file}"></script>`;
	}).join("\n")}
	</body>
	</html>
`;
}

app.get("/admin", (req, res) => {
	res.send(html("Admin", ["panels-admin-style.css"], ["admin-panel.js"]));
});

app.get("/user", (req, res) => {
	res.send(html("Settings", ["panels-user-style.css"], ["user-panel.js"]));
});

app.use("/assets", express.static("../assets/"));

if (process.env.NODE_ENV != "development") {
	console.log("adding static asset route");
	app.use(express.static("../assets/dist"));
}

module.exports = app;