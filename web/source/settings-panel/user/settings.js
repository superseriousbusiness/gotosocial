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
const formFields = require("../lib/form-fields");
const user = require("../redux/reducers/user").actions;

const Languages = require("../components/languages");
const Submit = require("../components/submit");

module.exports = function UserSettings() {
	const dispatch = Redux.useDispatch();
	const account = Redux.useSelector(state => state.user.settings);

	const { onTextChange, onCheckChange } = formFields(dispatch, user.setSettingsVal, account);

	const [errorMsg, setError] = React.useState("");
	const [statusMsg, setStatus] = React.useState("");

	function submit() {
		setStatus("PATCHing");
		setError("");
		return Promise.try(() => {
			return dispatch(api.user.updateSettings());
		}).then(() => {
			setStatus("Saved!");
		}).catch((e) => {
			setError(e.message);
			setStatus("");
		});
	}

	return (
		<div className="user-settings">
			<h1>Post settings</h1>
			<div className="labelselect">
				<label htmlFor="language">Default post language</label>
				<select id="language" autoComplete="language" value={account.source.language.toUpperCase()} onChange={onTextChange("source.language")}>
					<Languages />
				</select>
			</div>
			<div className="labelselect">
				<label htmlFor="privacy">Default post privacy</label>
				<select id="privacy" value={account.source.privacy} onChange={onTextChange("source.privacy")}>
					<option value="private">Private / followers-only)</option>
					<option value="unlisted">Unlisted</option>
					<option value="public">Public</option>
				</select>
				<a href="https://docs.gotosocial.org/en/latest/user_guide/posts/#privacy-settings" target="_blank" className="moreinfolink" rel="noreferrer">Learn more about post privacy settings (opens in a new tab)</a>
			</div>
			<div className="labelselect">
				<label htmlFor="format">Default post format</label>
				<select id="format" value={account.source.format} onChange={onTextChange("source.format")}>
					<option value="plain">Plain (default)</option>
					<option value="markdown">Markdown</option>
				</select>
				<a href="https://docs.gotosocial.org/en/latest/user_guide/posts/#input-types" target="_blank" className="moreinfolink" rel="noreferrer">Learn more about post format settings (opens in a new tab)</a>
			</div>				
			<div className="labelcheckbox">
				<label htmlFor="sensitive">Mark my posts as sensitive by default</label>
				<input id="sensitive" type="checkbox" checked={account.source.sensitive} onChange={onCheckChange("source.sensitive")}/>
			</div>
			<Submit onClick={submit} label="Save post settings" errorMsg={errorMsg} statusMsg={statusMsg}/>
		</div>
	);
};