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

const query = require("../../../lib/query");
const useFormSubmit = require("../../../lib/form/submit");

const {
	TextArea,
	Select,
} = require("../../../components/form/inputs");

const MutationButton = require("../../../components/form/mutation-button");

const { Error } = require("../../../components/error");
const ExportFormatTable = require("./export-format-table");

module.exports = function ImportExportForm({ form, submitParse, parseResult }) {
	const [submitExport, exportResult] = useFormSubmit(form, query.useExportDomainListMutation());

	function fileChanged(e) {
		const reader = new FileReader();
		reader.onload = function (read) {
			form.domains.value = read.target.result;
			submitParse();
		};
		reader.readAsText(e.target.files[0]);
	}

	React.useEffect(() => {
		if (exportResult.isSuccess) {
			form.domains.setter(exportResult.data);
		}
		/* eslint-disable-next-line react-hooks/exhaustive-deps */
	}, [exportResult]);

	return (
		<>
			<h1>Import / Export suspended domains</h1>
			<p>
				This page can be used to import and export lists of domains to suspend.
				Exports can be done in various formats, with varying functionality and support in other software.
				Imports will automatically detect what format is being processed.
			</p>
			<ExportFormatTable />
			<div className="import-export">
				<TextArea
					field={form.domains}
					label="Domains"
					placeholder={`google.com\nfacebook.com`}
					rows={8}
				/>

				<div className="button-grid">
					<MutationButton
						label="Import"
						type="button"
						onClick={() => submitParse()}
						result={parseResult}
						showError={false}
					/>
					<label className="button">
						Import file
						<input
							type="file"
							className="hidden"
							onChange={fileChanged}
							accept="application/json,text/plain,text/csv"
						/>
					</label>
					<b /> {/* grid filler */}
					<MutationButton
						label="Export"
						type="button"
						onClick={() => submitExport("export")}
						result={exportResult} showError={false}
					/>
					<MutationButton label="Export to file" type="button" onClick={() => submitExport("export-file")} result={exportResult} showError={false} />
					<div className="export-file">
						<span>
							as
						</span>
						<Select
							field={form.exportType}
							options={<>
								<option value="plain">Text</option>
								<option value="json">JSON</option>
								<option value="csv">CSV</option>
							</>}
						/>
					</div>
				</div>

				{parseResult.error && <Error error={parseResult.error} />}
				{exportResult.error && <Error error={exportResult.error} />}
			</div>
		</>
	);
};