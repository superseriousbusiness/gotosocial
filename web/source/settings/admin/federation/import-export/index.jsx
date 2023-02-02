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
const { Switch, Route, Redirect, useLocation } = require("wouter");

const query = require("../../../lib/query");

const {
	useTextInput,
} = require("../../../lib/form");

const useFormSubmit = require("../../../lib/form/submit");

const ProcessImport = require("./process");
const ImportExportForm = require("./form");

const baseUrl = "/settings/admin/federation/import-export";

module.exports = function ImportExport() {
	const form = {
		domains: useTextInput("domains"),
		exportType: useTextInput("exportType", { defaultValue: "plain", dontReset: true })
	};

	const [submitParse, parseResult] = useFormSubmit(form, query.useProcessDomainListMutation());

	const [_location, setLocation] = useLocation();

	return (
		<Switch>
			<Route path={`${baseUrl}/process`}>
				{parseResult.isSuccess ? (
					<>
						<h1>
							<span className="button" onClick={() => {
								parseResult.reset();
								setLocation(baseUrl);
							}}>
								&lt; back
							</span> Confirm import:
						</h1>
						<ProcessImport
							list={parseResult.data}
						/>
					</>
				) : <Redirect to={baseUrl} />}
			</Route>

			<Route>
				{!parseResult.isSuccess ? (
					<ImportExportForm
						form={form}
						submitParse={submitParse}
						parseResult={parseResult}
					/>
				) : <Redirect to={`${baseUrl}/process`} />}
			</Route>
		</Switch>
	);
};