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
const ReactDom = require("react-dom");

const createPanel = require("../lib/panel");

const Basic = require("./basic");
const Posts = require("./posts");
const Security = require("./security");

require("../base.css");
require("./style.css");

function UserPanel({oauth}) {
	const [account, setAccount] = React.useState({});
	const [allowCustomCSS, setAllowCustomCSS] = React.useState(false);
	const [errorMsg, setError] = React.useState("");
	const [statusMsg, setStatus] = React.useState("Fetching user info");

	React.useEffect(() => {

	}, [oauth, setAllowCustomCSS, setError, setStatus]);

	React.useEffect(() => {
		Promise.try(() => {
			return oauth.apiRequest("/api/v1/instance", "GET");
		}).then((json) => {
			setAllowCustomCSS(json.configuration.accounts.allow_custom_css);
			Promise.try(() => {
				return oauth.apiRequest("/api/v1/accounts/verify_credentials", "GET");
			}).then((json) => {
				setAccount(json);
			}).catch((e) => {
				setError(e.message);
				setStatus("");
			});
		}).catch((e) => {
			setError(e.message);
			setStatus("");
		});

	}, [oauth, setAllowCustomCSS, setAccount, setError, setStatus]);

	return (
		<React.Fragment>
			<div>
				<button className="logout" onClick={oauth.logout}>Log out of settings panel</button>
			</div>
			<Basic oauth={oauth} account={account} allowCustomCSS={allowCustomCSS}/>
			<Posts oauth={oauth} account={account}/>
			<Security oauth={oauth}/>
		</React.Fragment>
	);
}

createPanel("GoToSocial User Panel", ["read write"], UserPanel);