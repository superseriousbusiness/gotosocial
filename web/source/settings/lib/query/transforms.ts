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
 * Map a list of items into an object.
 * 
 * In the following example, a list of DomainPerms like the following:
 * 
 * ```json
 * [
 *   {
 *     "domain": "example.org",
 *     "public_comment": "aaaaa!!"
 *   },
 *   {
 *     "domain": "another.domain",
 *     "public_comment": "they are poo"
 *   }
 * ]
 * ```
 * 
 * Would be converted into an Object like the following:
 * 
 * ```json
 * {
 *   "example.org": {
 *     "domain": "example.org",
 *     "public_comment": "aaaaa!!"
 *   },
 *   "another.domain": {
 *     "domain": "another.domain",
 *     "public_comment": "they are poo"
 *   },
 * }
 * ```
 * 
 * If you pass a non-array type into this function it
 * will be converted into an array first, as a treat.
 * 
 * @example
 * const extended = gtsApi.injectEndpoints({
 *   endpoints: (build) => ({
 *     getDomainBlocks: build.query<MappedDomainPerms, void | null>({
 *       query: () => ({
 *         url: `/api/v1/admin/domain_blocks`
 *       }),
 *     transformResponse: listToKeyedObject<DomainPerm>("domain"),
 *   }),
 * });
 */
export function listToKeyedObject<T>(key: keyof T) {
	return (list: T[] | T): { [_ in keyof T]: T } => {
		// Ensure we're actually
		// dealing with an array.
		if (!Array.isArray(list)) {
			list = [list];
		}
		
		const entries = list.map((entry) => [entry[key], entry]); 
		return Object.fromEntries(entries);
	};
}
