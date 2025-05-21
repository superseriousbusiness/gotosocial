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
import { useInstanceV2Query } from "../../../lib/query/gts-api";
import Loading from "../../../components/loading";
import { InstanceV2 } from "../../../lib/types/instance";

export default function Instance() {
	// Load instance v2 data.
	const {
		data,
		isFetching,
		isLoading,
	} = useInstanceV2Query();
	
	if (isFetching || isLoading) {
		return <Loading />;
	}

	if (data === undefined) {
		throw "could not fetch instance v2";
	}



	return (
		<>

		</>
	);
}

function InstanceInfo({ instance }: { instance: InstanceV2 }) {
	return (
		<dl className="info-list">
			<div className="info-list-entry">
				<dt>Version:</dt>
				<dd>{instance.version}</dd>
			</div>

			<div className="info-list-entry">
				<dt>Version:</dt>
				<dd>{instance.version}</dd>
			</div>
		</dl>
	);
}
