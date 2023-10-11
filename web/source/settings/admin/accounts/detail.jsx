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

const React = require("react");
const { useRoute, Redirect } = require("wouter");

const query = require("../../lib/query");

const FormWithData = require("../../lib/form/form-with-data");

const { useBaseUrl } = require("../../lib/navigation/util");
const FakeProfile = require("../../components/fake-profile");
const MutationButton = require("../../components/form/mutation-button");

const useFormSubmit = require("../../lib/form/submit").default;
const { useValue, useTextInput } = require("../../lib/form");
const { TextInput } = require("../../components/form/inputs");

module.exports = function AccountDetail({ }) {
	const baseUrl = useBaseUrl();

	let [_match, params] = useRoute(`${baseUrl}/:accountId`);

	if (params?.accountId == undefined) {
		return <Redirect to={baseUrl} />;
	} else {
		return (
			<div className="account-detail">
				<h1>
					Account Details
				</h1>
				<FormWithData
					dataQuery={query.useGetAccountQuery}
					queryArg={params.accountId}
					DataForm={AccountDetailForm}
				/>
			</div>
		);
	}
};

function AccountDetailForm({ data: account }) {
	let content;
	if (account.suspended) {
		content = (
			<h2 className="error">Account is suspended.</h2>
		);
	} else {
		content = <ModifyAccount account={account} />;
	}

	return (
		<>
			<FakeProfile {...account} />

			{content}
		</>
	);
}

function ModifyAccount({ account }) {
	const form = {
		id: useValue("id", account.id),
		reason: useTextInput("text", {})
	};

	const [modifyAccount, result] = useFormSubmit(form, query.useActionAccountMutation());

	return (
		<form onSubmit={modifyAccount}>
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
	);
}