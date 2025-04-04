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
import EmailChange from "./email";
import PasswordChange from "./password";
import TwoFactor from "./twofactor";
import { useInstanceV1Query } from "../../../lib/query/gts-api";
import Loading from "../../../components/loading";
import { useUserQuery } from "../../../lib/query/user";

export default function Account() {
	// Load instance data.
	const {
		data: instance,
		isFetching: isFetchingInstance,
		isLoading: isLoadingInstance
	} = useInstanceV1Query();
	
	// Load user data.
	const {
		data: user,
		isFetching: isFetchingUser,
		isLoading: isLoadingUser
	} = useUserQuery();

	if (
		(isFetchingInstance || isLoadingInstance) ||
		(isFetchingUser || isLoadingUser)
	) {
		return <Loading />;
	}

	if (user === undefined) {
		throw "could not fetch user";
	}

	if (instance === undefined) {
		throw "could not fetch instance";
	}
	
	return (
		<>
			<h1>Account Settings</h1>
			<EmailChange
				oidcEnabled={instance.configuration.oidc_enabled}
				user={user}
			/>
			<PasswordChange
				oidcEnabled={instance.configuration.oidc_enabled}
			/>
			<TwoFactor
				oidcEnabled={instance.configuration.oidc_enabled}
				twoFactorEnabledAt={user.two_factor_enabled_at}
			/>
		</>
	);
}

