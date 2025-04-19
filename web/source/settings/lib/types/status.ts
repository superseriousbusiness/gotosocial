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

import { Account } from "./account";
import { CustomEmoji } from "./custom-emoji";

export interface Status {
	id: string;
	created_at: string;
	edited_at: string | null;
	in_reply_to_id: string | null;
	in_reply_to_account_id: string | null;
	sensitive: boolean;
	spoiler_text: string;
	visibility: string;
	language: string;
	uri: string;
	url: string;
	replies_count: number;
	reblogs_count: number;
	favourites_count: number;
	favourited: boolean;
	reblogged: boolean;
	muted: boolean;
	bookmarked: boolean;
	pinned: boolean;
	content: string,
	reblog: Status | null,
	account: Account,
	media_attachments: MediaAttachment[],
	mentions: [];
	tags: [];
	emojis: CustomEmoji[];
	card: null;
	poll: null;
}

export interface MediaAttachment {
	id: string;
	type: string;
	url: string;
	text_url: string;
	preview_url: string;
	remote_url: string | null;
	preview_remote_url: string | null;
	meta: MediaAttachmentMeta;
	description: string;
	blurhash: string;
}

interface MediaAttachmentMeta {
	original: {
		width: number;
		height: number;
		size: string;
		aspect: number;
	},
	small: {
		width: number;
		height: number;
		size: string;
		aspect: number;
	},
	focus: {
		x: number;
		y: number;
	}
}
