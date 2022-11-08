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

const React = require("react");
const Redux = require("react-redux");

const Submit = require("../components/submit");

const api = require("../lib/api");
const submit = require("../lib/submit");

const adminActions = require("../redux/reducers/instances").actions;

const {
	TextInput,
	TextArea,
	File
} = require("../components/form-fields").formFields(adminActions.setAdminSettingsVal, (state) => state.instances.adminSettings);

module.exports = function AdminSettings() {
	const dispatch = Redux.useDispatch();
	const instance = Redux.useSelector(state => state.instances.adminSettings);

	const [errorMsg, setError] = React.useState("");
	const [statusMsg, setStatus] = React.useState("");

	const updateSettings = submit(
		() => dispatch(api.admin.updateInstance()),
		{setStatus, setError}
	);

	return (
		<div>
			<h1>Instance Settings</h1>
			<TextInput
				id="title"
				name="Title"
				placeHolder="My GoToSocial instance"
			/>

			<div className="file-upload">
				<h3>Instance thumbnail</h3>
				<div>
					<img className="preview avatar" src={instance.thumbnail} alt={instance.thumbnail ? `Thumbnail image for the instance` : "No instance thumbnail image set"} />
					<File 
						id="thumbnail"
						fileType="image/*"
					/>
				</div>
			</div>

			<TextInput
				id="thumbnail_description"
				name="Instance thumbnail description"
				placeHolder="A cute little picture of a smiling sloth."
			/>

			<TextArea
				id="short_description"
				name="Short description"
				placeHolder="A small testing instance for the GoToSocial alpha."
			/>
			<TextArea
				id="description"
				name="Full description"
				placeHolder="A small testing instance for the GoToSocial alpha."
			/>

			<TextInput
				id="contact_account.username"
				name="Contact user (local account username)"
				placeHolder="admin"
			/>
			<TextInput
				id="email"
				name="Contact email"
				placeHolder="admin@example.com"
			/>

			<TextArea
				id="terms"
				name="Terms & Conditions"
				placeHolder=""
			/>

			<Submit onClick={updateSettings} label="Save" errorMsg={errorMsg} statusMsg={statusMsg} />
		</div>
	);
};