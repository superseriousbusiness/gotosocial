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

/**
 * OAuthToken represents a response
 * to an OAuth token request.
 */
export interface OAuthAccessToken {
	/**
	 * Most likely to be 'Bearer'
	 * but may be something else.
	 */
	token_type: string;
	/**
	 * The actual token. Can be passed in to
	 * authenticate further requests using the
	 * Authorization header and the token type.
	 */
	access_token: string;
}

export interface OAuthApp {
	client_id: string;
	client_secret: string;
}

export interface OAuthAccessTokenRequestBody {
	client_id: string;
	client_secret: string;
	redirect_uri: string;
	grant_type: string;
	code: string;
}
