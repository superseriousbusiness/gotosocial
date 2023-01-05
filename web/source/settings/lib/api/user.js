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

const Promise = require("bluebird");

const user = require("../../redux/reducers/user").actions;

module.exports = function ({ apiCall, getChanges }) {
	function updateCredentials(selector, keys) {
		return function (dispatch, getState) {
			return Promise.try(() => {
				const state = selector(getState());

				const update = getChanges(state, keys);

				return dispatch(apiCall("PATCH", "/api/v1/accounts/update_credentials", update, "form"));
			}).then((account) => {
				return dispatch(user.setAccount(account));
			});
		};
	}

	return {
		fetchAccount: function fetchAccount() {
			return function (dispatch, _getState) {
				return Promise.try(() => {
					return dispatch(apiCall("GET", "/api/v1/accounts/verify_credentials"));
				}).then((account) => {
					return dispatch(user.setAccount(account));
				});
			};
		},

		updateProfile: function updateProfile() {
			const formKeys = ["display_name", "locked", "source", "custom_css", "source.note", "enable_rss"];
			const renamedKeys = {
				"source.note": "note"
			};
			const fileKeys = ["header", "avatar"];

			return updateCredentials((state) => state.user.profile, {formKeys, renamedKeys, fileKeys});
		},

		updateSettings: function updateProfile() {
			const formKeys = ["source"];

			return updateCredentials((state) => state.user.settings, {formKeys});
		}
	};
};