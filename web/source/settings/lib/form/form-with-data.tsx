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

import { Error } from "../../components/error";
import Loading from "../../components/loading";
import { NoArg } from "../types/query";
import { FormWithDataQuery } from "./types";

export interface FormWithDataProps {
	dataQuery: FormWithDataQuery,
	DataForm: ({ data, ...props }) => React.JSX.Element,
	queryArg?: any,
}

/**
 * Wrap Form component inside component that fires the RTK Query call, so Form
 * will only be rendered when data is available to generate form-fields for.
 */
export default function FormWithData({ dataQuery, DataForm, queryArg, ...props }: FormWithDataProps) {
	if (!queryArg) {
		queryArg = NoArg;
	}

	// Trigger provided query.
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
		return <DataForm data={data} {...props} />;
	}
}
