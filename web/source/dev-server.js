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
	app.use(express.static("../assets/bundled"));
}

module.exports = app;