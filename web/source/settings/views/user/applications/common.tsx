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

import React, { useMemo } from "react";
import { App } from "../../../lib/types/application";
import { useStore } from "react-redux";
import { RootState } from "../../../redux/store";

export const useAppWebsite = (app: App) => {
	return useMemo(() => {
		if (!app.website) {
			return "";
		}
    
		try {
			// Try to parse nicely and return link.
			const websiteURL = new URL(app.website);
			const websiteURLStr = websiteURL.toString();
			return (
				<a
					href={websiteURLStr}
					target="_blank"
					rel="nofollow noreferrer noopener"
				>{websiteURLStr}</a>
			);
		} catch {
			// Fall back to returning string.
			return app.website;
		}
	}, [app.website]);
};

export const useCreated = (app: App) => {
	return useMemo(() => {
		const createdAt = new Date(app.created_at);
		return <time dateTime={app.created_at}>{createdAt.toDateString()}</time>;
	}, [app.created_at]);
};

export const useRedirectURIs= (app: App) => {
	return useMemo(() => {
		const length = app.redirect_uris.length;
		if (length === 1)  {
			return app.redirect_uris[0];
		}
    
		return app.redirect_uris.map((redirectURI, i) => {
			return i === 0 ? <>{redirectURI}</> : <><br/>{redirectURI}</>;
		});
    
	}, [app.redirect_uris]);
};

export const useCallbackURL = () => {
	const state = useStore().getState() as RootState;
	const instanceUrl = state.login.instanceUrl;
	if (instanceUrl === undefined) {
		throw "instanceUrl undefined";
	}

	return useMemo(() => {
		const url = new URL(instanceUrl);
		if (url === null) {
			throw "redirectURI null";
		}
		url.pathname = "/settings/user/applications/callback";
		return url.toString(); 
	}, [instanceUrl]);
};
