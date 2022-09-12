/*
	GoToSocial
	Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
const d = require("dotty");

const user = require("../../redux/reducers/user").actions;

module.exports = function ({ apiCall }) {
	function updateCredentials(selector, {formKeys=[], renamedKeys=[], fileKeys=[]}) {
		return function (dispatch, getState) {
			return Promise.try(() => {
				const state = selector(getState());

				const update = {};

				formKeys.forEach((key) => {
					d.put(update, key, d.get(state, key));
				});

				renamedKeys.forEach(([sendKey, intKey]) => {
					d.put(update, sendKey, d.get(state, intKey));
				});

				fileKeys.forEach((key) => {
					let file = d.get(state, `${key}File`);
					if (file != undefined) {
						d.put(update, key, file);
					}
				});

				console.log(update);

				return dispatch(apiCall("PATCH", "/api/v1/accounts/update_credentials", update, "form"));
			}).then((account) => {
				console.log(account);
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
			const formKeys = ["display_name", "locked", "source"];
			const renamedKeys = [["note", "source.note"]];
			const fileKeys = ["header", "avatar"];

			return updateCredentials((state) => state.user.profile, {formKeys, renamedKeys, fileKeys});
		},
		updateSettings: function updateProfile() {
			const formKeys = ["source"];

			return updateCredentials((state) => state.user.settings, {formKeys});
		}
	};
};