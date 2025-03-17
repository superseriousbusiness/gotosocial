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

import { PayloadAction, createSlice } from "@reduxjs/toolkit";
import { OAuthApp, OAuthAccessToken } from "../lib/types/oauth";

export interface LoginState {
	instanceUrl?: string;
	current: "none" | "awaitingcallback" | "loggedin" | "loggedout";
	expectingRedirect: boolean;
	/**
	 * Token stored in easy-to-use format.
	 * Will look something like:
	 * "Authorization: Bearer BLAHBLAHBLAH"
	 */
	token?: string;
	app?: OAuthApp;
}

const initialState: LoginState = {
	current: 'none',
	expectingRedirect: false,
};

export const loginSlice = createSlice({
	name: "login",
	initialState: initialState,
	reducers: {
		authorize: (_state, action: PayloadAction<LoginState>) => {
			// Overrides state with payload.
			return action.payload;
		},
		setToken: (state, action: PayloadAction<OAuthAccessToken>) => {
			// Mark us as logged
			// in by storing token.
			state.token = `${action.payload.token_type} ${action.payload.access_token}`;
			state.current = "loggedin";
		},
		remove: (state) => {
			// Mark us as logged
			// out by clearing auth.
			delete state.token;
			delete state.app;
			state.current = "loggedout";
		}
	}
});

export const {
	authorize,
	setToken,
	remove,
} = loginSlice.actions;
