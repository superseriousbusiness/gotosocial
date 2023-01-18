/*
	GoToSocial
	Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

const React = require("react");

function ErrorFallback({ error, resetErrorBoundary }) {
	return (
		<div className="error">
			<p>
				{"An error occured, please report this on the "}
				<a href="https://github.com/superseriousbusiness/gotosocial/issues">GoToSocial issue tracker</a>
				{" or "}
				<a href="https://matrix.to/#/#gotosocial-help:superseriousbusiness.org">Matrix support room</a>.
				<br />Include the details below:
			</p>
			<pre>
				{error.name}: {error.message}
			</pre>
			<pre>
				{error.stack}
			</pre>
			<p>
				<button onClick={resetErrorBoundary}>Try again</button> or <a href="">refresh the page</a>
			</p>
		</div>
	);
}

function Error({ error }) {
	/* eslint-disable-next-line no-console */
	console.error("Rendering error:", error);
	let message;

	if (error.data != undefined) { // RTK Query error with data
		if (error.status) {
			message = (<>
				<b>{error.status}:</b> {error.data.error}
				{error.data.error_description &&
					<p>
						{error.data.error_description}
					</p>
				}
			</>);
		} else {
			message = error.data.error;
		}
	} else if (error.name != undefined || error.type != undefined) { // JS error
		message = (<>
			<b>{error.type && error.name}:</b> {error.message}
		</>);
	} else if (error.status && typeof error.error == "string") {
		message = (<>
			<b>{error.status}:</b> {error.error}
		</>);
	} else {
		message = error.message ?? error;
	}

	return (
		<div className="error">
			{message}
		</div>
	);
}

module.exports = { ErrorFallback, Error };