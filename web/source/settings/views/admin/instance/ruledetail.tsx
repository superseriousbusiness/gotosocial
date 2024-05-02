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
import { Redirect, useParams } from "wouter";
import { useBaseUrl } from "../../../lib/navigation/util";
import { useValue, useTextInput } from "../../../lib/form";
import useFormSubmit from "../../../lib/form/submit";
import { TextArea } from "../../../components/form/inputs";
import MutationButton from "../../../components/form/mutation-button";
import BackButton from "../../../components/back-button";
import Loading from "../../../components/loading";
import { useDeleteInstanceRuleMutation, useInstanceRulesQuery, useUpdateInstanceRuleMutation } from "../../../lib/query/admin";
import { Error } from "../../../components/error";

export default function InstanceRuleDetail() {
	const baseUrl = useBaseUrl();
	const params: { ruleId: string } = useParams();
	
	const { data: rules, isLoading, isError, error } = useInstanceRulesQuery();
	if (isLoading) {
		return <Loading />;
	} else if (isError) {
		return <Error error={error} />;
	}

	if (rules === undefined) {
		throw "undefined rules";
	}

	return (
		<>
			<BackButton to={`~${baseUrl}/rules`} />
			<EditInstanceRuleForm rule={rules[params.ruleId]} />
		</>
	);
}

function EditInstanceRuleForm({ rule }) {
	const baseUrl = useBaseUrl();
	const form = {
		id: useValue("id", rule.id),
		rule: useTextInput("text", { defaultValue: rule.text })
	};

	const [submitForm, result] = useFormSubmit(form, useUpdateInstanceRuleMutation());

	const [deleteRule, deleteResult] = useDeleteInstanceRuleMutation({ fixedCacheKey: rule.id });

	if (result.isSuccess || deleteResult.isSuccess) {
		return (
			<Redirect to={`~${baseUrl}/rules`} />
		);
	}

	return (
		<div className="rule-detail">
			<form onSubmit={submitForm}>
				<TextArea
					field={form.rule}
				/>

				<div className="action-buttons row">
					<MutationButton
						label="Save"
						showError={false}
						result={result}
						disabled={!form.rule.hasChanged()}
					/>

					<MutationButton
						disabled={false}
						type="button"
						onClick={() => deleteRule(rule.id)}
						label="Delete"
						className="button danger"
						showError={false}
						result={deleteResult}
					/>
				</div>

				{result.error && <Error error={result.error} />}
				{deleteResult.error && <Error error={deleteResult.error} />}
			</form>
		</div>
	);
}
