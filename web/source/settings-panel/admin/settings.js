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

const Submit = require("../components/submit");

const api = require("../lib/api");
const formFields = require("../lib/form-fields");
const adminActions = require("../redux/reducers/instances").actions;

module.exports = function AdminSettings() {
	const dispatch = Redux.useDispatch();
	const instance = Redux.useSelector(state => state.instances.adminSettings);

	const { onTextChange, onCheckChange, onFileChange } = formFields(dispatch, adminActions.setAdminSettingsVal, instance);

	const [errorMsg, setError] = React.useState("");
	const [statusMsg, setStatus] = React.useState("");

	function submit() {
		setStatus("PATCHing");
		setError("");
		return Promise.try(() => {
			return dispatch(api.user.updateProfile());
		}).then(() => {
			setStatus("Saved!");
		}).catch((e) => {
			setError(e.message);
			setStatus("");
		});
	}

	// function removeFile(name) {
	// 	return function(e) {
	// 		e.preventDefault();
	// 		dispatch(user.setProfileVal([name, ""]));
	// 		dispatch(user.setProfileVal([`${name}File`, ""]));
	// 	};
	// }

	return (
		<div className="user-profile">
			<h1>Instance Settings</h1>

			<Submit onClick={submit} label="Save" errorMsg={errorMsg} statusMsg={statusMsg} />
		</div>
	);
};