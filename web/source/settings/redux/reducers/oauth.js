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

const {createSlice} = require("@reduxjs/toolkit");

module.exports = createSlice({
	name: "oauth",
	initialState: {
		loginState: 'none',
	},
	reducers: {
		setInstance: (state, {payload}) => {
			state.instance = payload;
		},
		setRegistration: (state, {payload}) => {
			state.registration = payload;
		},
		setLoginState: (state, {payload}) => {
			state.loginState = payload;
		},
		login: (state, {payload}) => {
			state.token = `${payload.token_type} ${payload.access_token}`;
			state.loginState = "login";
		},
		remove: (state, {_payload}) => {
			delete state.token;
			delete state.registration;
			delete state.isAdmin;
			state.loginState = "none";
		},
		setAdmin: (state, {payload}) => {
			state.isAdmin = payload;
		}
	}
});