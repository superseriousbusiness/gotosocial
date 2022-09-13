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

const instance = require("../../redux/reducers/instances").actions;

module.exports = function ({ apiCall, getChanges }) {
	return {
		updateInstance: function updateInstance() {
			return function (dispatch, getState) {
				return Promise.try(() => {
					const state = getState().instances.adminSettings;

					const update = getChanges(state, {
						formKeys: ["title", "short_description", "description", "contact_username", "email", "terms"],
						// fileKeys: ["avatar", "header"]
					});

					return dispatch(apiCall("PATCH", "/api/v1/instance", update, "form"));
				}).then((data) => {
					return dispatch(instance.setInstanceInfo(data));
				});
			};
		}
	};
};