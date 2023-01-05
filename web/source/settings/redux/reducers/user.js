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

const { createSlice } = require("@reduxjs/toolkit");
const d = require("dotty");

module.exports = createSlice({
	name: "user",
	initialState: {
		profile: {},
		settings: {}
	},
	reducers: {
		setAccount: (state, { payload }) => {
			payload.source = payload.source ?? {};
			payload.source.language = payload.source.language.toUpperCase() ?? "EN";
			payload.source.status_format = payload.source.status_format ?? "plain";
			payload.source.sensitive = payload.source.sensitive ?? false;

			state.profile = payload;
			// /user/settings only needs a copy of the 'source' obj
			state.settings = {
				source: payload.source
			};
		},
		setProfileVal: (state, { payload: [key, val] }) => {
			d.put(state.profile, key, val);
		},
		setSettingsVal: (state, { payload: [key, val] }) => {
			d.put(state.settings, key, val);
		}
	}
});