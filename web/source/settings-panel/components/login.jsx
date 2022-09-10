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

const { setInstance } = require("../redux/reducers/instances").actions;
const { updateInstance, updateRegistration } = require("../lib/api");

module.exports = function Login() {
	const dispatch = Redux.useDispatch();
	const [ instanceField, setInstanceField ] = React.useState("");
	const [ errorMsg, setErrorMsg ] = React.useState();
	const instanceFieldRef = React.useRef("");

	React.useEffect(() => {
		// check if current domain runs an instance
		Promise.try(() => {
			console.log("trying", window.location.origin);
			return dispatch(updateInstance(window.location.origin));
		}).then((json) => {
			if (instanceFieldRef.current.length == 0) { // user hasn't started typing yet
				dispatch(setInstance(json.uri));
				instanceFieldRef.current = json.uri;
				setInstanceField(json.uri);
			}
		}).catch((e) => {
			console.log("Current domain does not host a valid instance: ", e);
		});
	}, []);

	function tryInstance() {
		Promise.try(() => {
			return dispatch(updateInstance(instanceFieldRef.current)).catch((e) => {
				// TODO: clearer error messages for common errors
				console.log(e);
				throw e;
			});
		}).then((instance) => {
			// return dispatch(updateRegistration);
		}).catch((e) => {
			setErrorMsg(
				<>
					<b>{e.type}</b>
					<span>{e.message}</span>
				</>
			);
		});
	}

	function updateInstanceField(e) {
		if (e.key == "Enter") {
			tryInstance(instanceField);
		} else {
			setInstanceField(e.target.value);
			instanceFieldRef.current = e.target.value;
		}
	}

	return (
		<section className="login">
			<h1>OAUTH Login:</h1>
			<form onSubmit={(e) => e.preventDefault()}>
				<label htmlFor="instance">Instance: </label>
				<input value={instanceField} onChange={updateInstanceField} id="instance"/>
				{errorMsg && 
				<div className="error">
					{errorMsg}
				</div>
				}
				<button onClick={tryInstance}>Authenticate</button>
			</form>
		</section>
	);
};