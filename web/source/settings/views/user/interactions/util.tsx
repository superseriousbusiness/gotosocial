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

import { useMemo } from "react";

import sanitize from "sanitize-html";
import { compile, HtmlToTextOptions } from "html-to-text";
import { Status } from "../../../lib/types/status";

// Options for converting HTML statuses
// to plaintext representations.
const convertOptions: HtmlToTextOptions = {
	selectors: [
		// Don't fancy format links, just use their text value.
		{ selector: 'a', options: { ignoreHref: true } },
	]
};
const convertHTML = compile(convertOptions);

/**
 * Convert input status to plaintext representation.
 * @param status 
 * @returns 
 */
export function useContent(status: Status | undefined): string {
	return useMemo(() => {
		if (!status) {
			return "";
		}
		
		if (status.content.length === 0) {
			return "[no content set]";
		} else {
			// HTML has already been through
			// the instance sanitizer by now,
			// but do it again just in case.
			const content = sanitize(status.content);
			
			// Return plaintext of sanitized HTML.
			return convertHTML(content);
		}
	}, [status]);
}

export function useVerbed(type: "favourite" | "reply" | "reblog"): string {
	return useMemo(() => {
		switch (type) {
			case "favourite":
				return "liked";
			case "reply":
				return "replied to";
			case "reblog":
				return "boosted";
		}
	}, [type]);
}

export function useNoun(type: "favourite" | "reply" | "reblog"): string {
	return useMemo(() => {
		switch (type) {
			case "favourite":
				return "Like";
			case "reply":
				return "Reply";
			case "reblog":
				return "Boost";
		}
	}, [type]);
}

export function useIcon(type: "favourite" | "reply" | "reblog"): string {
	return useMemo(() => {
		switch (type) {
			case "favourite":
				return "fa-star";
			case "reply":
				return "fa-reply";
			case "reblog":
				return "fa-retweet";
		}
	}, [type]);
}
