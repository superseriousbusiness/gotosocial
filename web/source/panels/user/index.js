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

const oauthLib = require("../../lib/oauth.js");
const Auth = require("./auth");
const Basic = require("./basic");
const Posts = require("./posts");
const Security = require("./security");

require("../base.css");
require("./style.css");

function App() {
	const [oauth, setOauth] = React.useState();
	const [hasAuth, setAuth] = React.useState(false);
	const [oauthState, setOauthState] = React.useState(localStorage.getItem("oauth"));

	React.useEffect(() => {
		let state = localStorage.getItem("oauth");
		if (state != undefined) {
			state = JSON.parse(state);
			let restoredOauth = oauthLib(state.config, state);
			Promise.try(() => {
				return restoredOauth.callback();
			}).then(() => {
				setAuth(true);
			});
			setOauth(restoredOauth);
		}
	}, [setAuth, setOauth]);

	if (!hasAuth && oauth && oauth.isAuthorized()) {
		setAuth(true);
	}

	if (oauth && oauth.isAuthorized()) {
		return <UserPanel oauth={oauth} />;
	} else if (oauthState != undefined) {
		return "processing oauth...";
	} else {
		return <Auth setOauth={setOauth} />;
	}
}

function UserPanel({oauth}) {
   const [account, setAccount] = React.useState({});
	const [errorMsg, setError] = React.useState("");
	const [statusMsg, setStatus] = React.useState("Fetching user info");

   React.useEffect(() => {
      Promise.try(() => {
			return oauth.apiRequest("/api/v1/accounts/verify_credentials", "GET");
		}).then((json) => {
			setAccount(json);
		}).catch((e) => {
			setError(e.message);
			setStatus("");
		});
   }, [oauth, setAccount, setError, setStatus])

	return (
		<React.Fragment>
			<div>
				<button className="logout" onClick={oauth.logout}>Log out of settings panel</button>
			</div>
            <Basic oauth={oauth} account={account}/>
			<Posts oauth={oauth} account={account}/>
			<Security oauth={oauth}/>
		</React.Fragment>
	);
}

ReactDom.render(<App/>, document.getElementById("root"));
