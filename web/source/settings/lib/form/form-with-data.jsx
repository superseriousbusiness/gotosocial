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
const { Error } = require("../../components/error");

const Loading = require("../../components/loading");

// Wrap Form component inside component that fires the RTK Query call,
// so Form will only be rendered when data is available to generate form-fields for
module.exports = function FormWithData({ dataQuery, DataForm, queryArg, ...formProps }) {
	const { data, isLoading, isError, error } = dataQuery(queryArg);

	if (isLoading) {
		return (
			<div>
				<Loading />
			</div>
		);
	} else if (isError) {
		return (
			<Error error={error} />
		);
	} else {
		return <DataForm data={data} {...formProps} />;
	}
};