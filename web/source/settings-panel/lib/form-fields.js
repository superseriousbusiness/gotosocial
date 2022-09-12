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

const d = require("dotty");

module.exports = function(dispatch, setter, obj) {
	return {
		onTextChange: function (key) {
			return function (e) {
				dispatch(setter([key, e.target.value]));
			};
		},
	
		onCheckChange: function (key) {
			return function (e) {
				dispatch(setter([key, e.target.checked]));
			};
		},
	
		onFileChange: function (key) {
			return function (e) {
				let old = d.get(obj, key);
				if (old != undefined) {
					URL.revokeObjectURL(old); // no error revoking a non-Object URL as provided by instance
				}
				let file = e.target.files[0];
				let objectURL = URL.createObjectURL(file);
				dispatch(setter([key, objectURL]));
				dispatch(setter([`${key}File`, file]));
			};
		}
	};
};
