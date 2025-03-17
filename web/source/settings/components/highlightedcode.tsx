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

import React, { useEffect, useRef } from "react";

// Used for syntax highlighting of json result.
import Prism from "../../frontend/prism";

export function HighlightedCode({ code, lang }: { code: string, lang: string }) {
	const ref = useRef<HTMLElement | null>(null);
	useEffect(() => {
		if (ref.current) {
			Prism.highlightElement(ref.current);
		}
	}, []);

	// Prism takes control of the `pre` so wrap
	// the whole thing in a div that we control.
	return (
		<div className="prism-highlighted">
			<pre>
				<code ref={ref} className={`language-${lang}`}>
					{code}
				</code>
			</pre>
		</div>
	);
}
