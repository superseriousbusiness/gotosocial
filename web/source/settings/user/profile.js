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

const Submit = require("../components/submit");

const api = require("../lib/api");
const user = require("../redux/reducers/user").actions;
const submit = require("../lib/submit");

const FakeProfile = require("../components/fake-profile");
const { formFields } = require("../components/form-fields");

const {
	TextInput,
	TextArea,
	Checkbox,
	File
} = formFields(user.setProfileVal, (state) => state.user.profile);

module.exports = function UserProfile() {
	const dispatch = Redux.useDispatch();
	const instance = Redux.useSelector(state => state.instances.current);

	const allowCustomCSS = instance.configuration.accounts.allow_custom_css;

	const [errorMsg, setError] = React.useState("");
	const [statusMsg, setStatus] = React.useState("");

	const saveProfile = submit(
		() => dispatch(api.user.updateProfile()),
		{setStatus, setError}
	);

	return (
		<div className="user-profile">
			<h1>Profile</h1>
			<div className="overview">
				<FakeProfile/>
				<div className="files">
					<div>
						<h3>Header</h3>
						<File
							id="header"
							fileType="image/*"
						/>
					</div>
					<div>
						<h3>Avatar</h3>
						<File
							id="avatar"
							fileType="image/*"
						/>
					</div>
				</div>
			</div>
			<TextInput
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
			<Submit onClick={saveProfile} label="Save profile info" errorMsg={errorMsg} statusMsg={statusMsg} />
		</div>
	);
};