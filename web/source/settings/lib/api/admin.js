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
const isValidDomain = require("is-valid-domain");

const instance = require("../../redux/reducers/instances").actions;
const admin = require("../../redux/reducers/admin").actions;

module.exports = function ({ apiCall, getChanges }) {
	const adminAPI = {
		updateInstance: function updateInstance() {
			return function (dispatch, getState) {
				return Promise.try(() => {
					const state = getState().instances.adminSettings;

					const update = getChanges(state, {
						formKeys: ["title", "short_description", "description", "contact_account.username", "email", "terms", "thumbnail_description"],
						renamedKeys: {
							"email": "contact_email",
							"contact_account.username": "contact_username"
						},
						fileKeys: ["thumbnail"]
					});

					return dispatch(apiCall("PATCH", "/api/v1/instance", update, "form"));
				}).then((data) => {
					return dispatch(instance.setInstanceInfo(data));
				});
			};
		},

		fetchDomainBlocks: function fetchDomainBlocks() {
			return function (dispatch, _getState) {
				return Promise.try(() => {
					return dispatch(apiCall("GET", "/api/v1/admin/domain_blocks"));
				}).then((data) => {
					return dispatch(admin.setBlockedInstances(data));
				});
			};
		},

		updateDomainBlock: function updateDomainBlock(domain) {
			return function (dispatch, getState) {
				return Promise.try(() => {
					const state = getState().admin.newInstanceBlocks[domain];
					const update = getChanges(state, {
						formKeys: ["domain", "obfuscate", "public_comment", "private_comment"],
					});

					return dispatch(apiCall("POST", "/api/v1/admin/domain_blocks", update, "form"));
				}).then((block) => {
					return Promise.all([
						dispatch(admin.newDomainBlock([domain, block])),
						dispatch(admin.setDomainBlock([domain, block]))
					]);
				});
			};
		},

		getEditableDomainBlock: function getEditableDomainBlock(domain) {
			return function (dispatch, getState) {
				let data = getState().admin.blockedInstances[domain];
				return dispatch(admin.newDomainBlock([domain, data]));
			};
		},

		bulkDomainBlock: function bulkDomainBlock() {
			return function (dispatch, getState) {
				let invalidDomains = [];
				let success = 0;

				return Promise.try(() => {
					const state = getState().admin.bulkBlock;
					let list = state.list;
					let domains;

					let fields = getChanges(state, {
						formKeys: ["obfuscate", "public_comment", "private_comment"]
					});

					let defaultDate = new Date().toUTCString();
					
					if (list[0] == "[") {
						domains = JSON.parse(state.list);
					} else {
						domains = list.split("\n").map((line_) => {
							let line = line_.trim();
							if (line.length == 0) {
								return null;
							}

							if (!isValidDomain(line, {wildcard: true, allowUnicode: true})) {
								invalidDomains.push(line);
								return null;
							}

							return {
								domain: line,
								created_at: defaultDate,
								...fields
							};
						}).filter((a) => a != null);
					}

					if (domains.length == 0) {
						return;
					}

					const update = {
						domains: new Blob([JSON.stringify(domains)], {type: "application/json"})
					};

					return dispatch(apiCall("POST", "/api/v1/admin/domain_blocks?import=true", update, "form"));
				}).then((blocks) => {
					if (blocks != undefined) {
						return Promise.each(blocks, (block) => {
							success += 1;
							return dispatch(admin.setDomainBlock([block.domain, block]));
						});
					}
				}).then(() => {
					return {
						success,
						invalidDomains
					};
				});
			};
		},

		removeDomainBlock: function removeDomainBlock(domain) {
			return function (dispatch, getState) {
				return Promise.try(() => {
					const id = getState().admin.blockedInstances[domain].id;
					return dispatch(apiCall("DELETE", `/api/v1/admin/domain_blocks/${id}`));
				}).then((removed) => {
					return dispatch(admin.removeDomainBlock(removed.domain));
				});
			};
		},

		mediaCleanup: function mediaCleanup(days) {
			return function (dispatch, _getState) {
				return Promise.try(() => {
					return dispatch(apiCall("POST", `/api/v1/admin/media_cleanup?remote_cache_days=${days}`));
				});
			};
		},
	};
	return adminAPI;
};