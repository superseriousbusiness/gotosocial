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

import React from "react";
import { Link } from "wouter";
import { useInstanceRulesQuery, useAddInstanceRuleMutation } from "../../../lib/query/admin";
import { useBaseUrl } from "../../../lib/navigation/util";
import { useTextInput } from "../../../lib/form";
import useFormSubmit from "../../../lib/form/submit";
import { TextArea } from "../../../components/form/inputs";
import MutationButton from "../../../components/form/mutation-button";
import { InstanceRule, MappedRules } from "../../../lib/types/rules";
import FormWithData from "../../../lib/form/form-with-data";

export default function InstanceRules() {
	return (
		<>
			<h1>Instance Rules</h1>
			<FormWithData
				dataQuery={useInstanceRulesQuery}
				DataForm={InstanceRulesForm}
			/>
		</>
	);
}

function InstanceRulesForm({ data: rules }: { data: MappedRules }) {
	const baseUrl = useBaseUrl();
	const newRule = useTextInput("text");

	const [submitForm, result] = useFormSubmit({ newRule }, useAddInstanceRuleMutation(), {
		changedOnly: true,
		onFinish: () => newRule.reset()
	});

	return (
		<form onSubmit={submitForm} className="new-rule">
			<ol className="instance-rules">
				{Object.values(rules).map((rule: InstanceRule) => (
					<Link key={"link-"+rule.id} className="rule" to={`~${baseUrl}/rules/${rule.id}`}>
						<li key={rule.id}>
							<h2>{rule.text} <i className="fa fa-pencil edit-icon" /></h2>
						</li>
						<span>{new Date(rule.created_at).toLocaleString()}</span>
					</Link>
				))}
			</ol>
			<TextArea
				field={newRule}
				label="New instance rule"
			/>
			<MutationButton
				disabled={newRule.value === undefined || newRule.value.length === 0}
				label="Add rule"
				result={result}
			/>
		</form>
	);
}
