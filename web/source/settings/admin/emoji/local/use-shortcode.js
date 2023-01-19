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

const React = require("react");

const query = require("../../../lib/query");
const { useTextInput } = require("../../../lib/form");

const shortcodeRegex = /^[a-z0-9_]+$/;

module.exports = function useShortcode() {
	const {
		data: emoji = []
	} = query.useListEmojiQuery({ filter: "domain:local" });

	const emojiCodes = React.useMemo(() => {
		return new Set(emoji.map((e) => e.shortcode));
	}, [emoji]);

	return useTextInput("shortcode", {
		validator: function validateShortcode(code) {
			// technically invalid, but hacky fix to prevent validation error on page load
			if (code == "") { return ""; }

			if (emojiCodes.has(code)) {
				return "Shortcode already in use";
			}

			if (code.length < 2 || code.length > 30) {
				return "Shortcode must be between 2 and 30 characters";
			}

			if (code.toLowerCase() != code) {
				return "Shortcode must be lowercase";
			}

			if (!shortcodeRegex.test(code)) {
				return "Shortcode must only contain lowercase letters, numbers, and underscores";
			}

			return "";
		}
	});
};