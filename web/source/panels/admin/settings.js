"use strict";

const Promise = require("bluebird");
const React = require("react");

module.exports = function Settings({oauth}) {
	const [info, setInfo] = React.useState({});
	const [errorMsg, setError] = React.useState("");
	const [statusMsg, setStatus] = React.useState("Fetching instance info");

	React.useEffect(() => {
		Promise.try(() => {
			return oauth.apiRequest("/api/v1/instance", "GET");
		}).then((json) => {
			setInfo(json);
		}).catch((e) => {
			setError(e.message);
			setStatus("");
		});
	}, []);

	function submit() {
		setStatus("PATCHing");
		setError("");
		return Promise.try(() => {
			let formDataInfo = new FormData();
			Object.entries(info).forEach(([key, val]) => {
				if (key == "contact_account") {
					key = "contact_username";
					val = val.username;
				}
				if (key == "email") {
					key = "contact_email";
				}
				if (typeof val != "object") {
					formDataInfo.append(key, val);
				}
			});
			return oauth.apiRequest("/api/v1/instance", "PATCH", formDataInfo, "form");
		}).then((json) => {
			setStatus("Config saved");
			console.log(json);
		}).catch((e) => {
			setError(e.message);
			setStatus("");
		});
	}

	return (
		<section className="info login">
			<h1>Instance Information <button onClick={submit}>Save</button></h1>
			<div className="error accent">
				{errorMsg}
			</div>
			<div>
				{statusMsg}
			</div>
			<form onSubmit={(e) => e.preventDefault()}>
				{editableObject(info)}
			</form>
		</section>
	);
};

function editableObject(obj, path=[]) {
	const readOnlyKeys = ["uri", "version", "urls_streaming_api", "stats"];
	const hiddenKeys = ["contact_account_", "urls"];
	const explicitShownKeys = ["contact_account_username"];
	const implementedKeys = "title, contact_account_username, email, short_description, description, terms, avatar, header".split(", ");

	let listing = Object.entries(obj).map(([key, val]) => {
		let fullkey = [...path, key].join("_");

		if (
			hiddenKeys.includes(fullkey) ||
			hiddenKeys.includes(path.join("_")+"_") // also match just parent path
		) {
			if (!explicitShownKeys.includes(fullkey)) {
				return null;
			}
		}

		if (Array.isArray(val)) {
			// FIXME: handle this
		} else if (typeof val == "object") {
			return (<React.Fragment key={fullkey}>
				{editableObject(val, [...path, key])}
			</React.Fragment>);
		} 

		let isImplemented = "";
		if (!implementedKeys.includes(fullkey)) {
			isImplemented = " notImplemented";
		}

		let isReadOnly = (
			readOnlyKeys.includes(fullkey) ||
			readOnlyKeys.includes(path.join("_")) ||
			isImplemented != ""
		);

		let label = key.replace(/_/g, " ");
		if (path.length > 0) {
			label = `\u00A0`.repeat(4 * path.length) + label;
		}

		let inputProps;
		let changeFunc;
		if (val === true || val === false) {
			inputProps = {
				type: "checkbox",
				defaultChecked: val,
				disabled: isReadOnly
			};
			changeFunc = (e) => e.target.checked;
		} else if (val.length != 0 && !isNaN(val)) {
			inputProps = {
				type: "number",
				defaultValue: val,
				readOnly: isReadOnly
			};
			changeFunc = (e) => e.target.value;
		} else {
			inputProps = {
				type: "text",
				defaultValue: val,
				readOnly: isReadOnly
			};
			changeFunc = (e) => e.target.value;
		}

		function setRef(element) {
			if (element != null) {
				element.addEventListener("change", (e) => {
					obj[key] = changeFunc(e);
				});
			}
		}

		return (
			<React.Fragment key={fullkey}>
				<label htmlFor={key} className="capitalize">{label}</label>
				<div className={isImplemented}>
					<input className={isImplemented} ref={setRef} {...inputProps} />
				</div>
			</React.Fragment>
		);
	});
	return (
		<React.Fragment>
			{path != "" &&
				<><b>{path}:</b> <span id="filler"></span></>
			}
			{listing}
		</React.Fragment>
	);
}