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
const React = require("react");
const Redux = require("react-redux");

const api = require("../lib/api");
const adminActions = require("../redux/reducers/admin").actions;

module.exports = function CustomEmoji() {
	return (
		<>
			<h1>Custom Emoji</h1>
			<div>
				<EmojiOverview/>
			</div>
			<div>
				<h2>Upload</h2>
			</div>
		</>
	);
};

function EmojiOverview() {
	const dispatch = Redux.useDispatch();
	const emoji = Redux.useSelector((state) => state.admin.emoji);
	console.log(emoji);

	React.useEffect(() => {
		dispatch(api.admin.fetchCustomEmoji());
	}, []);

	return (
		<>

		</>
	);
}