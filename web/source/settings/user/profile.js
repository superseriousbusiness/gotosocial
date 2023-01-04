/*
	GoToSocial
	Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

const query = require("../lib/query");

const {
	useTextInput
} = require("../components/form");

const FakeProfile = require("../components/fake-profile");
const syncpipe = require("syncpipe");
const MutationButton = require("../components/mutation-button");

module.exports = function UserProfile() {
	const allowCustomCSS = Redux.useSelector(state => state.instances.current.configuration.accounts.allow_custom_css);
	const profile = Redux.useSelector(state => state.user.profile);

	/*
		User profile update form keys
		- bool bot
		- bool locked
		- string display_name
		- string note
		- file avatar
		- file header
		- string source[privacy]
		- bool source[sensitive]
		- string source[language]
		- string source[status_format]
		- bool enable_rss
		- string custom_css (if enabled)
	*/

	const form = {
		display_name: useTextInput("displayName", {defaultValue: profile.display_name})
	};

	const [result, submitForm] = useFormSubmit(form, query.useUpdateCredentialsMutation());

	return (
		<form className="user-profile" onSubmit={submitForm}>
			<h1>Profile</h1>
			<div className="overview">
				{/* <FakeProfile/> */}
				<div className="files">
					<div>
						<h3>Header</h3>
						{/* <File
							id="header"
							fileType="image/*"
						/> */}
					</div>
					<div>
						<h3>Avatar</h3>
						{/* <File
							id="avatar"
							fileType="image/*"
						/> */}
					</div>
				</div>
			</div>
			<FormTextInput
				label="Name"
				placeHolder="A GoToSocial user"
				field={form.display_name}
			/>
			{/* <TextInput
				id="display_name"
				name="Name"
				placeHolder="A GoToSocial user"
			/>
			<TextArea
				id="source.note"
				name="Bio"
				placeHolder="Just trying out GoToSocial, my pronouns are they/them and I like sloths."
			/>
			<Checkbox
				id="locked"
				name="Manually approve follow requests"
			/>
			<Checkbox
				id="enable_rss"
				name="Enable RSS feed of Public posts"
			/>
			{ !allowCustomCSS ? null :
				<TextArea
					id="custom_css"
					name="Custom CSS"
					className="monospace"
				>
					<a href="https://docs.gotosocial.org/en/latest/user_guide/custom_css" target="_blank" className="moreinfolink" rel="noreferrer">Learn more about custom profile CSS (opens in a new tab)</a>
				</TextArea>
			}
			<Submit onClick={saveProfile} label="Save profile info" errorMsg={errorMsg} statusMsg={statusMsg} /> */}
			<MutationButton text="Save profile info" result={result}/>
		</form>
	);
};

function FormTextInput({label, placeHolder, field}) {
	let [onChange, _reset, {value, ref}] = field;

	return (
		<div className="form-field text">
			<label>
				{label}
				<input
					type="text"
					placeholder={placeHolder}
					{...{onChange, value, ref}}
				/>
			</label>
		</div>
	);
}

function useFormSubmit(form, [mutationQuery, result]) {
	return [
		result,
		function submitForm(e) {
			e.preventDefault();

			// transform the field definitions into an object with just their values 
			let updatedFields = 0;
			const mutationData = syncpipe(form, [
				(_) => Object.entries(_),
				(_) => _.map(([key, field]) => {
					let data = field[2]; // [onChange, reset, {}]
					if (data.hasChanged()) {
						return [key, data.value];
					} else {
						return null;
					}
				}),
				(_) => _.filter((value) => value != null),
				(_) => {
					updatedFields = _.length;
					return _;
				},
				(_) => Object.fromEntries(_)
			]);

			if (updatedFields > 0) {
				return mutationQuery(mutationData);
			}
		},
	];
}

// function useForm(formSpec) {
// 	const form = {};

// 	Object.entries(formSpec).forEach(([name, cfg]) => {
// 		const [useTypedInput, defaultValue] = cfg;

// 		form[name] = useTypedInput(name, );
// 	});

// 	form.submit = function submitForm() {

// 	};

// 	return form;
// }