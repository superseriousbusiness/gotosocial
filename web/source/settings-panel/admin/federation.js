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
const {Switch, Route, Link, useRoute} = require("wouter");

const Submit = require("../components/submit");

const api = require("../lib/api");
const adminActions = require("../redux/reducers/instances").actions;

const base = "/settings/admin/federation";

// const {
// 	TextInput,
// 	TextArea,
// 	File
// } = require("../components/form-fields").formFields(adminActions.setAdminSettingsVal, (state) => state.instances.adminSettings);

module.exports = function AdminSettings() {
	const dispatch = Redux.useDispatch();
	// const instance = Redux.useSelector(state => state.instances.adminSettings);
	const { blockedInstances } = Redux.useSelector(state => state.admin);

	const [errorMsg, setError] = React.useState("");
	const [statusMsg, setStatus] = React.useState("");

	const [loaded, setLoaded] = React.useState(false);

	React.useEffect(() => {
		if (blockedInstances != undefined) {
			setLoaded(true);
		} else {
			return Promise.try(() => {
				return dispatch(api.admin.fetchDomainBlocks());
			}).then(() => {
				setLoaded(true);
			});
		}
	}, []);

	if (!loaded) {
		return (
			<div>
				<h1>Federation</h1>
				Loading...
			</div>
		);
	}

	return (
		<div>
			<Switch>
				<Route path={`${base}/:domain`}>
					<InstancePage blockedInstances={blockedInstances}/>
				</Route>
				<InstanceOverview blockedInstances={blockedInstances} />
			</Switch>
		</div>
	);
};

function InstanceOverview({blockedInstances}) {
	return (
		<div>
			<h1>Federation</h1>
			{blockedInstances.map((entry) => {
				return (
					<Link key={entry.domain} to={`${base}/${entry.domain}`}>
						<a>{entry.domain}</a>
					</Link>
				);
			})}
		</div>
	);
}

function BackButton() {
	return (
		<Link to={base}>
			<a className="button">&lt; back</a>
		</Link>
	);
}

function InstancePage({blockedInstances}) {
	let [_match, {domain}] = useRoute(`${base}/:domain`);
	let [status, setStatus] = React.useState("");
	let [entry, setEntry] = React.useState(() => {
		let entry = blockedInstances.find((a) => a.domain == domain);
	
		if (entry == undefined) {
			setStatus(`No block entry found for ${domain}, but you can create one below:`);
			return {
				private_comment: ""
			};
		} else {
			return entry;
		}
	});

	return (
		<div>
			{status}
			<h1><BackButton/> Federation settings for: {domain}</h1>
			<div>{entry.private_comment}</div>
		</div>
	);
}