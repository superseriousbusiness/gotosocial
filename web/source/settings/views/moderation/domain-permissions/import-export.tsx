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

import { Switch, Route, Redirect, useLocation } from "wouter";
import { useProcessDomainPermissionsMutation } from "../../../lib/query/admin/domain-permissions/process";
import { useTextInput, useRadioInput } from "../../../lib/form";
import useFormSubmit from "../../../lib/form/submit";
import { ProcessImport } from "./process";
import ImportExportForm from "./form";

export default function ImportExport() {
	const form = {
		domains: useTextInput("domains"),
		exportType: useTextInput("exportType", {
			defaultValue: "plain",
			dontReset: true,
		}),
		permType: useRadioInput("permType", { 
			options: {
				block: "Domain blocks",
				allow: "Domain allows",
			}
		})
	};

	const [submitParse, parseResult] = useFormSubmit(form, useProcessDomainPermissionsMutation(), { changedOnly: false });
	const [_location, setLocation] = useLocation();

	return (
		<Switch>
			<Route path={"/process"}>
				{
					parseResult.isSuccess 
						? (
							<>
								<h1>
									<span
										className="button"
										onClick={() => {
											parseResult.reset();
											setLocation("");
										}}
									>
										&lt; back
									</span>
									&nbsp; Confirm {form.permType.value}s:
								</h1>
								<ProcessImport
									list={parseResult.data}
									permType={form.permType}
								/>
							</>
						)
						: <Redirect to={""} />
				}
			</Route>
			<Route>
				{
					parseResult.isSuccess
						? <Redirect to={"/process"} />
						: <ImportExportForm
							form={form}
							submitParse={submitParse}
							parseResult={parseResult}
						/>
				}
			</Route>
		</Switch>
	);
}
