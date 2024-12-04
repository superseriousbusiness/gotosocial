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

export interface InstanceV1 {
    uri:                    string;
    account_domain:         string;
    title:                  string;
    description:            string;
    description_text?:       string;
    short_description:      string;
    short_description_text?: string;
    custom_css:             string;
    email:                  string;
    version:                string;
    debug?:                 boolean;
    languages:              any[]; // TODO: define this
    registrations:          boolean;
    approval_required:      boolean;
    invites_enabled:        boolean;
    configuration:          InstanceConfiguration;
    urls:                   InstanceUrls;
    stats:                  InstanceStats;
    thumbnail:              string;
    contact_account:        Object; // TODO: define this.
    max_toot_chars:         number;
    rules:                  any[]; // TODO: define this
    terms?:                 string;
    terms_text?:             string;
}

export interface InstanceConfiguration {
    statuses:          InstanceStatuses;
    media_attachments: InstanceMediaAttachments;
    polls:             InstancePolls;
    accounts:          InstanceAccounts;
    emojis:            InstanceEmojis;
    oidc_enabled?:     boolean;
}

export interface InstanceAccounts {
    allow_custom_css:   boolean;
    max_featured_tags:  number;
    max_profile_fields: number;
}

export interface InstanceEmojis {
    emoji_size_limit: number;
}

export interface InstanceMediaAttachments {
    supported_mime_types:   string[];
    image_size_limit:       number;
    image_matrix_limit:     number;
    video_size_limit:       number;
    video_frame_rate_limit: number;
    video_matrix_limit:     number;
}

export interface InstancePolls {
    max_options:               number;
    max_characters_per_option: number;
    min_expiration:            number;
    max_expiration:            number;
}

export interface InstanceStatuses {
    max_characters:              number;
    max_media_attachments:       number;
    characters_reserved_per_url: number;
    supported_mime_types:        string[];
}

export interface InstanceStats {
    domain_count: number;
    status_count: number;
    user_count:   number;
}

export interface InstanceUrls {
    streaming_api: string;
}

