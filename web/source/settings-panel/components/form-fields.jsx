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
const Redux = require("react-redux");
const d = require("dotty");

function eventListeners(dispatch, setter, obj) {
	return {
		onTextChange: function (key) {
			return function (e) {
				dispatch(setter([key, e.target.value]));
			};
		},
		
		onCheckChange: function (key) {
			return function (e) {
				dispatch(setter([key, e.target.checked]));
			};
		},
		
		onFileChange: function (key) {
			return function (e) {
				let old = d.get(obj, key);
				if (old != undefined) {
					URL.revokeObjectURL(old); // no error revoking a non-Object URL as provided by instance
				}
				let file = e.target.files[0];
				let objectURL = URL.createObjectURL(file);
				dispatch(setter([key, objectURL]));
				dispatch(setter([`${key}File`, file]));
			};
		}
	};
}

function get(state, id) {
	let value;
	if (id.includes(".")) {
		value = d.get(state, id);
	} else {
		value = state[id];
	}
	return value;
}

// function removeFile(name) {
// 	return function(e) {
// 		e.preventDefault();
// 		dispatch(user.setProfileVal([name, ""]));
// 		dispatch(user.setProfileVal([`${name}File`, ""]));
// 	};
// }

module.exports = {
	formFields: function formFields(setter, selector) {
		function FormField({type, id, name, className="", placeHolder="", fileType="", children=null}) {
			const dispatch = Redux.useDispatch();
			let state = Redux.useSelector(selector);
			let {
				onTextChange,
				onCheckChange,
				onFileChange
			} = eventListeners(dispatch, setter, state);

			let field;
			let defaultLabel = true;
			if (type == "text") {
				field = <input type="text" id={id} value={get(state, id)} placeholder={placeHolder} className={className} onChange={onTextChange(id)}/>;
			} else if (type == "textarea") {
				field = <textarea type="text" id={id} value={get(state, id)} placeholder={placeHolder} className={className} onChange={onTextChange(id)}/>;
			} else if (type == "checkbox") {
				field = <input type="checkbox" id={id} checked={get(state, id)} className={className} onChange={onCheckChange(id)}/>;
			} else if (type == "file") {
				defaultLabel = false;
				let file = get(state, `${id}File`);
				field = (
					<>
						<label htmlFor={id} className="file-input button">Browse</label>
						<span>{file ? file.name : "no file selected"}</span>
						{/* <a onClick={removeFile("header")} href="#">remove</a> */}
						<input className="hidden" id={id} type="file" accept={fileType} onChange={onFileChange(id)} />
					</>
				);
			} else {
				defaultLabel = false;
				field = `unsupported FormField ${type}, this is a developer error`;
			}

			let label = <label htmlFor={id}>{name}</label>;
	
			return (
				<div className={`form-field ${type}`}>
					{defaultLabel ? label : null}
					{field}
					{children}
				</div>
			);
		}

		return {
			TextInput: function(props) {
				return <FormField type="text" {...props} />;
			},
	
			TextArea: function(props) {
				return <FormField type="textarea" {...props} />;
			},
	
			Checkbox: function(props) {
				return <FormField type="checkbox" {...props} />;
			},
	
			File: function(props) {
				return <FormField type="file" {...props} />;
			},
		};
	},

	eventListeners
};