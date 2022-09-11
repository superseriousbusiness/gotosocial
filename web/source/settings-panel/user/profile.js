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

module.exports = function UserProfile() {
	const dispatch = Redux.useDispatch();
	const account = Redux.useSelector(state => state.user.account);

	const [errorMsg, setError] = React.useState("");
	const [statusMsg, setStatus] = React.useState("");

	const [headerFile, setHeaderFile] = React.useState(undefined);
	const [avatarFile, setAvatarFile] = React.useState(undefined);

	const [displayName, setDisplayName] = React.useState("");
	const [bio, setBio] = React.useState("");
	const [locked, setLocked] = React.useState(false);

	React.useEffect(() => {

		setDisplayName(account.display_name);
		setBio(account.source ? account.source.note : "");
		setLocked(account.locked);
	}, []);

	const headerOnChange = (e) => {
		setHeaderFile(e.target.files[0]);
		// setHeaderSrc(URL.createObjectURL(e.target.files[0]));
	};

	const avatarOnChange = (e) => {
		setAvatarFile(e.target.files[0]);
		// setAvatarSrc(URL.createObjectURL(e.target.files[0]));
	};

	const submit = (e) => {
		e.preventDefault();

		setStatus("PATCHing");
		setError("");
		return Promise.try(() => {
			let payload = {
				display_name: displayName,
				note: bio,
				locked: locked
			};

			if (headerFile) {
				payload.header = headerFile;
			}

			if (avatarFile) {
				payload.avatar = avatarFile;
			}

			return dispatch(api.user.updateAccount(payload));
		}).then(() => {
			setStatus("Saved!");
		}).catch((e) => {
			setError(e.message);
			setStatus("");
		});
	};

	return (
		<div className="user-profile">
			<h1>Profile</h1>
			<div className="overview">
				<div className="profile">
        	<div className="headerimage">
						<img className="headerpreview" src={account.header} alt={account.header ? `header image for ${account.username}` : "None set"}/>
        	</div>
        	<div className="basic">
           	<div id="profile-basic-filler2"></div>
						<span className="avatar"><img className="avatarpreview" src={account.avatar} alt={account.avatar ? `avatar image for ${account.username}` : "None set"}/></span>
           	<div className="displayname">{account.display_name.trim().length > 0 ? account.display_name : account.username}</div>
           	<div className="username"><span>@{account.username}</span></div>
        	</div>
				</div>
				<div className="files">
					<div>
						<h3>Header</h3>
						<label htmlFor="header" className="file-input button">Browse…</label>
						<span>{headerFile ? headerFile.name : "no file selected"}</span>
					</div>
					<div>
						<h3>Avatar</h3>
						<label htmlFor="avatar" className="file-input button">Browse…</label>
						<span>{avatarFile ? avatarFile.name : "no file selected"}</span>
					</div>
				</div>
			</div>
			<div className="labelinput">
				<label htmlFor="displayname">Name</label>
				<input id="displayname" type="text" value={displayName} onChange={(e) => setDisplayName(e.target.value)} placeholder="A GoToSocial user"/>
			</div>
			<div className="labelinput">
				<label htmlFor="bio">Bio</label>
				<textarea id="bio" value={bio} onChange={(e) => setBio(e.target.value)} placeholder="Just trying out GoToSocial, my pronouns are they/them and I like sloths."/>
			</div>
			<div className="labelcheckbox">
				<label htmlFor="locked">Manually approve follow requests?</label>
				<input id="locked" type="checkbox" checked={locked} onChange={(e) => setLocked(e.target.checked)}/>
			</div>
			<Submit onClick={submit} label="Save profile info" errorMsg={errorMsg} statusMsg={statusMsg}/>
		</div>
	);
};