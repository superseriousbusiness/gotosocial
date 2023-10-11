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
const { Switch, Route, Link, Redirect, useRoute } = require("wouter");

const query = require("../../lib/query");
const FormWithData = require("../../lib/form/form-with-data");
const { useBaseUrl } = require("../../lib/navigation/util");

const { useValue, useTextInput } = require("../../lib/form");
const useFormSubmit = require("../../lib/form/submit").default;

const { TextArea } = require("../../components/form/inputs");
const MutationButton = require("../../components/form/mutation-button");

module.exports = function InstanceRulesData({ baseUrl }) {
	return (
		<FormWithData
			dataQuery={query.useInstanceRulesQuery}
			DataForm={InstanceRules}
			baseUrl={baseUrl}
		/>
	);
};

function InstanceRules({ baseUrl, data: rules }) {
	return (
		<Switch>
			<Route path={`${baseUrl}/:ruleId`}>
				<InstanceRuleDetail rules={rules} />
			</Route>
			<Route>
				<div>
					<h1>Instance Rules</h1>
					<div>
						<p>
							The rules for your instance are listed on the about page, and can be selected when submitting reports.
						</p>
					</div>
					<InstanceRuleList rules={rules} />
				</div>
			</Route>
		</Switch>
	);
}

function InstanceRuleList({ rules }) {
	const newRule = useTextInput("text", {});

	const [submitForm, result] = useFormSubmit({ newRule }, query.useAddInstanceRuleMutation(), {
		onFinish: () => newRule.reset()
	});

	return (
		<>
			<form onSubmit={submitForm} className="new-rule">
				<ol className="instance-rules">
					{Object.values(rules).map((rule) => (
						<InstanceRule key={rule.id} rule={rule} />
					))}
				</ol>
				<TextArea
					field={newRule}
					label="New instance rule"
				/>
				<MutationButton label="Add rule" result={result} />
			</form>
		</>
	);
}

function InstanceRule({ rule }) {
	const baseUrl = useBaseUrl();

	return (
		<Link to={`${baseUrl}/${rule.id}`}>
			<a className="rule">
				<li>
					<h2>{rule.text} <i className="fa fa-pencil edit-icon" /></h2>
				</li>
				<span>{new Date(rule.created_at).toLocaleString()}</span>
			</a>
		</Link>
	);
}

function InstanceRuleDetail({ rules }) {
	const baseUrl = useBaseUrl();
	let [_match, params] = useRoute(`${baseUrl}/:ruleId`);

	if (params?.ruleId == undefined || rules[params.ruleId] == undefined) {
		return <Redirect to={baseUrl} />;
	} else {
		return (
			<>
				<Link to={baseUrl}><a>&lt; go back</a></Link>
				<InstanceRuleForm rule={rules[params.ruleId]} />
			</>
		);
	}
}

function InstanceRuleForm({ rule }) {
	const baseUrl = useBaseUrl();
	const form = {
		id: useValue("id", rule.id),
		rule: useTextInput("text", { defaultValue: rule.text })
	};

	const [submitForm, result] = useFormSubmit(form, query.useUpdateInstanceRuleMutation());

	const [deleteRule, deleteResult] = query.useDeleteInstanceRuleMutation({ fixedCacheKey: rule.id });

	if (result.isSuccess || deleteResult.isSuccess) {
		return (
			<Redirect to={baseUrl} />
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