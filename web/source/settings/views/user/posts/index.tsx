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
import { useVerifyCredentialsQuery } from "../../../lib/query/oauth";
import Loading from "../../../components/loading";
import { Error } from "../../../components/error";
import BasicSettings from "./basic-settings";
import InteractionPolicySettings from "./interaction-policy-settings";

export default function PostSettings() {
	const {
		data: account,
		isLoading,
		isFetching,
		isError,
		error,
	} = useVerifyCredentialsQuery();

	if (isLoading || isFetching) {
		return <Loading />;
	}

	if (isError) {
		return <Error error={error} />;
	}

	return (
		<>
			<h1>Post Settings</h1>
			<BasicSettings account={account} />
			<InteractionPolicySettings />
		</>
	);
}
