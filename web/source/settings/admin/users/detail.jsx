/*
	GoToSocial
	Copyright (C) GoToSocial Authors admin@gotosocial.org
	SPDX-License-Identifier: AGPL-3.0-or-later

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
const { useRoute, Redirect } = require("wouter");

const query = require("../../lib/query");

const FormWithData = require("../../lib/form/form-with-data");

const { useBaseUrl } = require("../../lib/navigation/util");
const FakeProfile = require("../../components/fake-profile");
const MutationButton = require("../../components/form/mutation-button");

const useFormSubmit = require("../../lib/form/submit");
const { useValue, useTextInput } = require("../../lib/form");
const { TextInput } = require("../../components/form/inputs");

module.exports = function UserDetail({ }) {
	const baseUrl = useBaseUrl();

	let [_match, params] = useRoute(`${baseUrl}/:userId`);

	if (params?.userId == undefined) {
		return <Redirect to={baseUrl} />;
	} else {
		return (
			<div className="user-detail">
				<h1>
					User Details
				</h1>
				<FormWithData
					dataQuery={query.useGetUserQuery}
					queryArg={params.userId}
					DataForm={UserDetailForm}
				/>
			</div>
		);
	}
};

function UserDetailForm({ data: user }) {

	const form = {
		id: useValue("id", user.id),
		reason: useTextInput("text", {})
	};

	const [modifyUser, result] = useFormSubmit(form, query.useActionUserMutation());

	return (
		<div>
			<FakeProfile {...user} />

			{user.suspended &&
				<h2 style={{ color: "red" }}>NUKED FROM ORBIT</h2>
			}

			<form onSubmit={modifyUser}>
				<h2>Actions</h2>
				<TextInput
					field={form.reason}
					placeholder="Reason for this action"
				/>

				<div className="action-buttons">
					{/* <MutationButton
						label="Disable"
						name="disable"
						result={result}
					/>
					<MutationButton
						label="Silence"
						name="silence"
						result={result}
					/> */}
					<MutationButton
						label="Suspend"
						name="suspend"
						result={result}
					/>
				</div>
			</form>
		</div>
	);
}