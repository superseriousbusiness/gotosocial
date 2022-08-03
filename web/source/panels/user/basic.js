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
const Promise = require("bluebird");

module.exports = function Basic({oauth, account}) {
	const [errorMsg, setError] = React.useState("");
	const [statusMsg, setStatus] = React.useState("");

    const [headerFile, setHeaderFile] = React.useState(undefined);
    const [headerSrc, setHeaderSrc] = React.useState("");

    const [avatarFile, setAvatarFile] = React.useState(undefined);
    const [avatarSrc, setAvatarSrc] = React.useState("");

    const [displayName, setDisplayName] = React.useState("");
    const [bio, setBio] = React.useState("");
    const [locked, setLocked] = React.useState(false);

    React.useEffect(() => {
        setHeaderSrc(account.header);
        setAvatarSrc(account.avatar);

        setDisplayName(account.display_name);
        setBio(account.source ? account.source.note : "");
        setLocked(account.locked);
    }, [account, setHeaderSrc, setAvatarSrc, setDisplayName, setBio, setLocked]);

    const headerOnChange = (e) => {
        setHeaderFile(e.target.files[0]);
        setHeaderSrc(URL.createObjectURL(e.target.files[0]));
    }

    const avatarOnChange = (e) => {
        setAvatarFile(e.target.files[0]);
        setAvatarSrc(URL.createObjectURL(e.target.files[0]));
    }

	function submit() {
		setStatus("PATCHing");
		setError("");
		return Promise.try(() => {
			let formDataInfo = new FormData();

            if (headerFile) {
                formDataInfo.set("header", headerFile);
            }

            if (avatarFile) {
                formDataInfo.set("avatar", avatarFile);
            }

            formDataInfo.set("display_name", displayName);
            formDataInfo.set("note", bio);
            formDataInfo.set("locked", locked);

			return oauth.apiRequest("/api/v1/accounts/update_credentials", "PATCH", formDataInfo, "form");
		}).then((json) => {
			setStatus("Saved!");

            setHeaderSrc(json.header);
            setAvatarSrc(json.avatar);

            setDisplayName(json.display_name);
            setBio(json.source.note);
            setLocked(json.locked);
		}).catch((e) => {
			setError(e.message);
			setStatus("");
		});
	}

    return (
        <section className="basic">
            <h1>@{account.username}'s Profile Info</h1>
            <form>
                <div className="labelinput">
                    <label htmlFor="header">Header</label>
                    <img className="headerpreview" src={headerSrc} alt={headerSrc ? `header image for ${account.username}` : "None set"}/>
                    <input id="header" type="file" accept="image/*" onChange={headerOnChange}/>
                </div>
                <div className="labelinput">
                    <label htmlFor="avatar">Avatar</label>
                    <img className="avatarpreview" src={avatarSrc} alt={headerSrc ? `avatar image for ${account.username}` : "None set"}/>
                    <input id="avatar" type="file" accept="image/*" onChange={avatarOnChange}/>
                </div>
                <div className="labelinput">
                    <label htmlFor="displayname">Display Name</label>
                    <input id="displayname" type="text" value={displayName} onChange={(e) => setDisplayName(e.target.value)} />
                </div>
                <div className="labelinput">
                    <label htmlFor="bio">Bio</label>
                    <textarea id="bio" value={bio} onChange={(e) => setBio(e.target.value)} />
                </div>
                <div className="labelcheckbox">
                    <label htmlFor="locked">Manually approve follow requests</label>
                    <input id="locked" type="checkbox" checked={locked} onChange={(e) => setLocked(e.target.checked)}/>
                </div>
            </form>
            <div className="messagebutton">
                <button onClick={submit}>Save profile info</button>
                <div className="error accent">{errorMsg ? errorMsg : statusMsg}</div>
            </div>
        </section>
	);
}
