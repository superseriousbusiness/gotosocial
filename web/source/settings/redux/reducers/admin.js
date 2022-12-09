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

const { createSlice } = require("@reduxjs/toolkit");

function sortBlocks(blocks) {
	return blocks.sort((a, b) => { // alphabetical sort
		return a.domain.localeCompare(b.domain);
	});
}

function emptyBlock() {
	return {
		public_comment: "",
		private_comment: "",
		obfuscate: false
	};
}

module.exports = createSlice({
	name: "admin",
	initialState: {
		loadedBlockedInstances: false,
		blockedInstances: undefined,
		bulkBlock: {
			list: "",
			exportType: "plain",
			...emptyBlock()
		},
		newInstanceBlocks: {}
	},
	reducers: {
		setBlockedInstances: (state, { payload }) => {
			state.blockedInstances = {};
			sortBlocks(payload).forEach((entry) => {
				state.blockedInstances[entry.domain] = entry;
			});
			state.loadedBlockedInstances = true;
		},

		newDomainBlock: (state, { payload: [domain, data] }) => {
			if (data == undefined) {
				data = {
					new: true,
					domain,
					...emptyBlock()
				};
			}
			state.newInstanceBlocks[domain] = data;
		},

		setDomainBlock: (state, { payload: [domain, data = {}] }) => {
			state.blockedInstances[domain] = data;
		},

		removeDomainBlock: (state, {payload: domain}) => {
			delete state.blockedInstances[domain];
		},

		updateDomainBlockVal: (state, { payload: [domain, key, val] }) => {
			state.newInstanceBlocks[domain][key] = val;
		},

		updateBulkBlockVal: (state, { payload: [key, val] }) => {
			state.bulkBlock[key] = val;
		},

		resetBulkBlockVal: (state, { _payload }) => {
			state.bulkBlock = {
				list: "",
				exportType: "plain",
				...emptyBlock()
			};
		},

		exportToField: (state, { _payload }) => {
			state.bulkBlock.list = Object.values(state.blockedInstances).map((entry) => {
				return entry.domain;
			}).join("\n");
		}
	}
});