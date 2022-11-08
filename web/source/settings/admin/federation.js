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
const {Switch, Route, Link, Redirect, useRoute, useLocation} = require("wouter");
const fileDownload = require("js-file-download");

const { formFields } = require("../components/form-fields");

const api = require("../lib/api");
const adminActions = require("../redux/reducers/admin").actions;
const submit = require("../lib/submit");
const BackButton = require("../components/back-button");

const base = "/settings/admin/federation";

// const {
// 	TextInput,
// 	TextArea,
// 	File
// } = require("../components/form-fields").formFields(adminActions.setAdminSettingsVal, (state) => state.instances.adminSettings);

module.exports = function AdminSettings() {
	const dispatch = Redux.useDispatch();
	// const instance = Redux.useSelector(state => state.instances.adminSettings);
	const loadedBlockedInstances = Redux.useSelector(state => state.admin.loadedBlockedInstances);

	React.useEffect(() => {
		if (!loadedBlockedInstances ) {
			Promise.try(() => {
				return dispatch(api.admin.fetchDomainBlocks());
			});
		}
	}, [dispatch, loadedBlockedInstances]);

	if (!loadedBlockedInstances) {
		return (
			<div>
				<h1>Federation</h1>
				Loading...
			</div>
		);
	}

	return (
		<Switch>
			<Route path={`${base}/:domain`}>
				<InstancePageWrapped />
			</Route>
			<InstanceOverview />
		</Switch>
	);
};

function InstanceOverview() {
	const [filter, setFilter] = React.useState("");
	const blockedInstances = Redux.useSelector(state => state.admin.blockedInstances);
	const [_location, setLocation] = useLocation();

	function filterFormSubmit(e) {
		e.preventDefault();
		setLocation(`${base}/${filter}`);
	}

	return (
		<>
			<h1>Federation</h1>
			Here you can see an overview of blocked instances.

			<div className="instance-list">
				<h2>Blocked instances</h2>
				<form action={`${base}/view`} className="filter" role="search" onSubmit={filterFormSubmit}>
					<input name="domain" value={filter} onChange={(e) => setFilter(e.target.value)}/>
					<Link to={`${base}/${filter}`}><a className="button">Add block</a></Link>
				</form>
				<div className="list">
					{Object.values(blockedInstances).filter((a) => a.domain.startsWith(filter)).map((entry) => {
						return (
							<Link key={entry.domain} to={`${base}/${entry.domain}`}>
								<a className="entry nounderline">
									<span id="domain">
										{entry.domain}
									</span>
									<span id="date">
										{new Date(entry.created_at).toLocaleString()}
									</span>
								</a>
							</Link>
						);
					})}
				</div>
			</div>

			<BulkBlocking/>
		</>
	);
}

const Bulk = formFields(adminActions.updateBulkBlockVal, (state) => state.admin.bulkBlock);
function BulkBlocking() {
	const dispatch = Redux.useDispatch();
	const {bulkBlock, blockedInstances} = Redux.useSelector(state => state.admin);

	const [errorMsg, setError] = React.useState("");
	const [statusMsg, setStatus] = React.useState("");

	function importBlocks() {
		setStatus("Processing");
		setError("");
		return Promise.try(() => {
			return dispatch(api.admin.bulkDomainBlock());
		}).then(({success, invalidDomains}) => {
			return Promise.try(() => {
				return resetBulk();
			}).then(() => {
				dispatch(adminActions.updateBulkBlockVal(["list", invalidDomains.join("\n")]));

				let stat = "";
				if (success == 0) {
					return setError("No valid domains in import");
				} else if (success == 1) {
					stat = "Imported 1 domain";
				} else {
					stat = `Imported ${success} domains`;
				}

				if (invalidDomains.length > 0) {
					if (invalidDomains.length == 1) {
						stat += ", input contained 1 invalid domain.";
					} else {
						stat += `, input contained ${invalidDomains.length} invalid domains.`;
					}
				} else {
					stat += "!";
				}

				setStatus(stat);
			});
		}).catch((e) => {
			console.error(e);
			setError(e.message);
			setStatus("");
		});
	}

	function exportBlocks() {
		return Promise.try(() => {
			setStatus("Exporting");
			setError("");
			let asJSON = bulkBlock.exportType.startsWith("json");
			let _asCSV = bulkBlock.exportType.startsWith("csv");

			let exportList = Object.values(blockedInstances).map((entry) => {
				if (asJSON) {
					return {
						domain: entry.domain,
						public_comment: entry.public_comment
					};
				} else {
					return entry.domain;
				}
			});
			
			if (bulkBlock.exportType == "json") {
				return dispatch(adminActions.updateBulkBlockVal(["list", JSON.stringify(exportList)]));
			} else if (bulkBlock.exportType == "json-download") {
				return fileDownload(JSON.stringify(exportList), "block-export.json");
			} else if (bulkBlock.exportType == "plain") {
				return dispatch(adminActions.updateBulkBlockVal(["list", exportList.join("\n")]));
			}
		}).then(() => {
			setStatus("Exported!");
		}).catch((e) => {
			setError(e.message);
			setStatus("");
		});
	}

	function resetBulk(e) {
		if (e != undefined) {
			e.preventDefault();
		}
		return dispatch(adminActions.resetBulkBlockVal());
	}

	function disableInfoFields(props={}) {
		if (bulkBlock.list[0] == "[") {
			return {
				...props,
				disabled: true,
				placeHolder: "Domain list is a JSON import, input disabled"
			};
		} else {
			return props;
		}
	}

	return (
		<div className="bulk">
			<h2>Import / Export <a onClick={resetBulk}>reset</a></h2>
			<Bulk.TextArea
				id="list"
				name="Domains, one per line"
				placeHolder={`google.com\nfacebook.com`}
			/>

			<Bulk.TextArea
				id="public_comment"
				name="Public comment"
				inputProps={disableInfoFields({rows: 3})}
			/>

			<Bulk.TextArea
				id="private_comment"
				name="Private comment"
				inputProps={disableInfoFields({rows: 3})}
			/>

			<Bulk.Checkbox
				id="obfuscate"
				name="Obfuscate domains? "
				inputProps={disableInfoFields()}
			/>

			<div className="hidden">
				<Bulk.File
					id="json"
					fileType="application/json"
					withPreview={false}
				/>
			</div>

			<div className="messagebutton">
				<div>
					<button type="submit" onClick={importBlocks}>Import</button>
				</div>

				<div>
					<button type="submit" onClick={exportBlocks}>Export</button>

					<Bulk.Select id="exportType" name="Export type" options={
						<>
							<option value="plain">One per line in text field</option>
							<option value="json">JSON in text field</option>
							<option value="json-download">JSON file download</option>
							<option disabled value="csv">CSV in text field (glitch-soc)</option>
							<option disabled value="csv-download">CSV file download (glitch-soc)</option>
						</>
					}/>
				</div>
				<br/>
				<div>
					{errorMsg.length > 0 && 
						<div className="error accent">{errorMsg}</div>
					}
					{statusMsg.length > 0 &&
						<div className="accent">{statusMsg}</div>
					}
				</div>
			</div>
		</div>
	);
}

function InstancePageWrapped() {
	/* We wrap the component to generate formFields with a setter depending on the domain
		 if formFields() is used inside the same component that is re-rendered with their state,
		 inputs get re-created on every change, causing them to lose focus, and bad performance
	*/
	let [_match, {domain}] = useRoute(`${base}/:domain`);

	if (domain == "view") { // from form field submission
		let realDomain = (new URL(document.location)).searchParams.get("domain");
		if (realDomain == undefined) {
			return <Redirect to={base}/>;
		} else {
			domain = realDomain;
		}
	}

	function alterDomain([key, val]) {
		return adminActions.updateDomainBlockVal([domain, key, val]);
	}

	const fields = formFields(alterDomain, (state) => state.admin.newInstanceBlocks[domain]);

	return <InstancePage domain={domain} Form={fields} />;
}

function InstancePage({domain, Form}) {
	const dispatch = Redux.useDispatch();
	const entry = Redux.useSelector(state => state.admin.newInstanceBlocks[domain]);
	const [_location, setLocation] = useLocation();

	React.useEffect(() => {
		if (entry == undefined) {
			dispatch(api.admin.getEditableDomainBlock(domain));
		}
	}, [dispatch, domain, entry]);

	const [errorMsg, setError] = React.useState("");
	const [statusMsg, setStatus] = React.useState("");

	if (entry == undefined) {
		return "Loading...";
	}

	const updateBlock = submit(
		() => dispatch(api.admin.updateDomainBlock(domain)),
		{setStatus, setError}
	);

	const removeBlock = submit(
		() => dispatch(api.admin.removeDomainBlock(domain)),
		{setStatus, setError, startStatus: "Removing", successStatus: "Removed!", onSuccess: () => {
			setLocation(base);
		}}
	);

	return (
		<div>
			<h1><BackButton to={base}/> Federation settings for: {domain}</h1>
			{entry.new && "No stored block yet, you can add one below:"}

			<Form.TextArea
				id="public_comment"
				name="Public comment"
			/>

			<Form.TextArea
				id="private_comment"
				name="Private comment"
			/>

			<Form.Checkbox
				id="obfuscate"
				name="Obfuscate domain? "
			/>

			<div className="messagebutton">
				<button type="submit" onClick={updateBlock}>{entry.new ? "Add block" : "Save block"}</button>

				{!entry.new &&
					<button className="danger" onClick={removeBlock}>Remove block</button>
				}

				{errorMsg.length > 0 && 
					<div className="error accent">{errorMsg}</div>
				}
				{statusMsg.length > 0 &&
					<div className="accent">{statusMsg}</div>
				}
			</div>
		</div>
	);
}