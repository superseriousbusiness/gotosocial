"use strict";

const Promise = require("bluebird");
const React = require("react");
const ReactDom = require("react-dom");

const oauthLib = require("./oauth.js");
const Auth = require("./auth");
const Settings = require("./settings");
const Blocks = require("./blocks");

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
	}, []);

	if (!hasAuth && oauth && oauth.isAuthorized()) {
		setAuth(true);
	}

	if (oauth && oauth.isAuthorized()) {
		return <AdminPanel oauth={oauth} />;
	} else if (oauthState != undefined) {
		return "processing oauth...";
	} else {
		return <Auth setOauth={setOauth} />;
	}
}

function AdminPanel({oauth}) {
	/* 
		Features: (issue #78)
		- [ ] Instance information updating
			  GET /api/v1/instance PATCH /api/v1/instance
		- [ ] Domain block creation, viewing, and deletion
			  GET /api/v1/admin/domain_blocks
			  POST /api/v1/admin/domain_blocks
			  GET /api/v1/admin/domain_blocks/DOMAIN_BLOCK_ID, DELETE /api/v1/admin/domain_blocks/DOMAIN_BLOCK_ID
		- [ ] Blocklist import/export
			  GET /api/v1/admin/domain_blocks?export=true
			  POST json file as form field domains to /api/v1/admin/domain_blocks
	*/

	return (
		<React.Fragment>
			<Logout oauth={oauth}/>
			<Settings oauth={oauth} />
			<Blocks oauth={oauth}/>
		</React.Fragment>
	);
}

function Logout({oauth}) {
	return (
		<div>
			<button onClick={oauth.logout}>Logout</button>
		</div>
	);
}

ReactDom.render(<App/>, document.getElementById("root"));