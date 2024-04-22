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

import { CustomEmoji } from "./custom-emoji";

export interface AdminAccount {
	id: string,
	username: string,
	domain: string | null,
	created_at: string,
	email: string,
	ip: string | null,
	ips: [],
	locale: string,
	invite_request: string | null,
	role: any,
	confirmed: boolean,
	approved: boolean,
	disabled: boolean,
	silenced: boolean,
	suspended: boolean,
	created_by_application_id: string,
	account: Account,
}

export interface Account {
	id: string,
	username: string,
	acct: string,
	display_name: string,
	locked: boolean,
	discoverable: boolean,
	bot: boolean,
	created_at: string,
	note: string,
	url: string,
	avatar: string,
	avatar_static: string,
	header: string,
	header_static: string,
	followers_count: number,
	following_count: number,
	statuses_count: number,
	last_status_at: string,
	emojis: CustomEmoji[],
	fields: [],
	enable_rss: boolean,
	role: any,
}

export interface SearchAccountParams {
	origin?: "local" | "remote",
	status?: "active" | "pending" | "disabled" | "silenced" | "suspended",
	permissions?: "staff",
	username?: string,
	display_name?: string,
	by_domain?: string,
	email?: string,
	ip?: string,
	max_id?: string,
	since_id?: string,
	min_id?: string,
	limit?: number,
}

export interface HandleSignupParams {
	id: string,
	approve_or_reject: "approve" | "reject",
	private_comment?: string,
	message?: string,
	send_email?: boolean,
}
