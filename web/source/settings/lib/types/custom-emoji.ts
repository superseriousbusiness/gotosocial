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

export interface CustomEmoji {
	id?: string;
	shortcode: string;
	url: string;
	static_url: string;
	visible_in_picker: boolean;
	category?: string;
	disabled: boolean;
	updated_at: string;
	total_file_size: number;
	content_type: string;
	uri: string;
}

/**
 * Query parameters for GET to /api/v1/admin/custom_emojis.
 */
export interface ListEmojiParams {

}

/**
 * Result of searchItemForEmoji mutation.
 */
export interface EmojisFromItem {
	/**
	 * Type of the search item result.
	 */
	type: "statuses" | "accounts";
	/**
	 * Domain of the returned emojis.
	 */
	domain: string;
	/**
	 * Discovered emojis.
	 */
	list: CustomEmoji[];
}
