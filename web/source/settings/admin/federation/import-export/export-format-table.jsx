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

module.exports = function ExportFormatTable() {
	return (
		<table className="export-format-table">
			<thead>
				<tr>
					<th rowSpan={2} />
					<th colSpan={2}>Includes</th>
					<th colSpan={2}>Importable by</th>
				</tr>
				<tr>
					<th>Domain</th>
					<th>Public comment</th>
					<th>GoToSocial</th>
					<th>Mastodon</th>
				</tr>
			</thead>
			<tbody>
				<Format name="Text" info={[true, false, true, false]} />
				<Format name="JSON" info={[true, true, true, false]} />
				<Format name="CSV" info={[true, true, true, true]} />
			</tbody>
		</table>
	);
};

function Format({ name, info }) {
	return (
		<tr>
			<td><b>{name}</b></td>
			{info.map((b, key) => <td key={key} className="bool">{bool(b)}</td>)}
		</tr>
	);
}

function bool(val) {
	return (
		<>
			<i className={`fa fa-${val ? "check" : "times"}`} aria-hidden="true"></i>
			<span className="sr-only">{val ? "Yes" : "No"}</span>
		</>
	);
}