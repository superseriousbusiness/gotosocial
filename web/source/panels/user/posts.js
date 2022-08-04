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

const Languages = require("./languages");
const Submit = require("../../lib/submit");

module.exports = function Posts({oauth, account}) {
	const [errorMsg, setError] = React.useState("");
	const [statusMsg, setStatus] = React.useState("");

	const [language, setLanguage] = React.useState("");
	const [privacy, setPrivacy] = React.useState("");
	const [sensitive, setSensitive] = React.useState(false);

	React.useEffect(() => {
		if (account.source) {
			setLanguage(account.source.language.toUpperCase());
			setPrivacy(account.source.privacy);
			setSensitive(account.source.sensitive ? account.source.sensitive : false);
		}
        
	}, [account, setSensitive, setPrivacy]);

	const submit = (e) => {
		e.preventDefault();

		setStatus("PATCHing");
		setError("");
		return Promise.try(() => {
			let formDataInfo = new FormData();

			formDataInfo.set("source[language]", language);
			formDataInfo.set("source[privacy]", privacy);
			formDataInfo.set("source[sensitive]", sensitive);

			return oauth.apiRequest("/api/v1/accounts/update_credentials", "PATCH", formDataInfo, "form");
		}).then((json) => {
			setStatus("Saved!");
			setLanguage(json.source.language.toUpperCase());
			setPrivacy(json.source.privacy);
			setSensitive(json.source.sensitive ? json.source.sensitive : false);
		}).catch((e) => {
			setError(e.message);
			setStatus("");
		});
	};

	return (
		<section className="posts">
			<h1>Post Settings</h1>
			<form>
				<div className="labelselect">
					<label htmlFor="language">Default post language</label>
					<select id="language" autoComplete="language" value={language} onChange={(e) => setLanguage(e.target.value)}>
						<Languages />
					</select>
				</div>
				<div className="labelselect">
					<label htmlFor="privacy">Default post privacy</label>
					<select id="privacy" value={privacy} onChange={(e) => setPrivacy(e.target.value)}>
						<option value="private">Private / followers-only)</option>
						<option value="unlisted">Unlisted</option>
						<option value="public">Public</option>
					</select>
					<a href="https://docs.gotosocial.org/en/latest/user_guide/posts/#privacy-settings" target="_blank" className="moreinfolink" rel="noreferrer">Learn more about post privacy settings (opens in a new tab)</a>
				</div>
				<div className="labelcheckbox">
					<label htmlFor="sensitive">Mark my posts as sensitive by default</label>
					<input id="sensitive" type="checkbox" checked={sensitive} onChange={(e) => setSensitive(e.target.checked)}/>
				</div>
				<Submit onClick={submit} label="Save post settings" errorMsg={errorMsg} statusMsg={statusMsg}/>
			</form>
		</section>
	);
};
