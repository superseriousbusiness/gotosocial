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
const user = require("../redux/reducers/user").actions;

module.exports = function UserProfile() {
	const dispatch = Redux.useDispatch();
	const account = Redux.useSelector(state => state.user.profile);
	const instance = Redux.useSelector(state => state.instances.current);

	const allowCustomCSS = instance.configuration.accounts.allow_custom_css;

	const { onTextChange, onCheckChange, onFileChange } = formFields(dispatch, user.setProfileVal, account);

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
			<h1>Profile</h1>
			<div className="overview">
				<div className="profile">
					<div className="headerimage">
						<img className="headerpreview" src={account.header} alt={account.header ? `header image for ${account.username}` : "None set"} />
					</div>
					<div className="basic">
						<div id="profile-basic-filler2"></div>
						<span className="avatar"><img className="avatarpreview" src={account.avatar} alt={account.avatar ? `avatar image for ${account.username}` : "None set"} /></span>
						<div className="displayname">{account.display_name.trim().length > 0 ? account.display_name : account.username}</div>
						<div className="username"><span>@{account.username}</span></div>
					</div>
				</div>
				<div className="files">
					<div>
						<h3>Header</h3>
						<div className="picker">
							<label htmlFor="header" className="file-input button">Browse</label>
							<span>{account.headerFile ? account.headerFile.name : "no file selected"}</span>
						</div>
						{/* <a onClick={removeFile("header")} href="#">remove</a> */}
						<input className="hidden" id="header" type="file" accept="image/*" onChange={onFileChange("header")} />
					</div>
					<div>
						<h3>Avatar</h3>
						<div className="picker">
							<label htmlFor="avatar" className="file-input button">Browse</label>
							<span>{account.avatarFile ? account.avatarFile.name : "no file selected"}</span>
						</div>
						{/* <a onClick={removeFile("avatar")} href="#">remove</a> */}
						<input className="hidden" id="avatar" type="file" accept="image/*" onChange={onFileChange("avatar")} />
					</div>
				</div>
			</div>
			<div className="labelinput">
				<label htmlFor="displayname">Name</label>
				<input id="displayname" type="text" value={account.display_name} onChange={onTextChange("display_name")} placeholder="A GoToSocial user" />
			</div>
			<div className="labelinput">
				<label htmlFor="bio">Bio</label>
				<textarea id="bio" value={account.source.note} onChange={onTextChange("source.note")} placeholder="Just trying out GoToSocial, my pronouns are they/them and I like sloths." />
			</div>
			<div className="labelcheckbox">
				<label htmlFor="locked">Manually approve follow requests?</label>
				<input id="locked" type="checkbox" checked={account.locked} onChange={onCheckChange("locked")} />
			</div>
			{ !allowCustomCSS ? null :  
				<div className="labelinput">
					<label htmlFor="customcss">Custom CSS</label>
					<textarea className="mono" id="customcss" value={account.custom_css} onChange={onTextChange("custom_css")}/>
					<a href="https://docs.gotosocial.org/en/latest/user_guide/custom_css" target="_blank" className="moreinfolink" rel="noreferrer">Learn more about custom CSS (opens in a new tab)</a>
				</div>
			}
			<Submit onClick={submit} label="Save profile info" errorMsg={errorMsg} statusMsg={statusMsg} />
		</div>
	);
};