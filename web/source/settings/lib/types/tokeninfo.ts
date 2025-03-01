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

import { Links } from "parse-link-header";

export interface TokenInfo {
	id: string;
	created_at: string;
	last_used?: string;
	scope: string;
	application: {
		name: string;
		website?: string;
	};
}

/**
 * Parameters for GET to /api/v1/tokens.
 */
export interface SearchTokenInfoParams {
	/**
	 * If set, show only items older (ie., lower) than the given ID.
	 * Item with the given ID will not be included in response.
	 */
	max_id?: string;
	/**
	 * If set, show only items newer (ie., higher) than the given ID.
	 * Item with the given ID will not be included in response.
	 */
	since_id?: string;
	/**
	 * If set, show only items *immediately newer* than the given ID.
	 * Item with the given ID will not be included in response.
	 */
	min_id?: string;
	/**
	 * If set, limit returned items to this number.
	 * Else, fall back to GtS API defaults.
	 */
	limit?: number;
}

export interface SearchTokenInfoResp {
	tokens: TokenInfo[];
	links: Links | null;
}
