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
import { useTextInput } from "../../../../lib/form";
import { useLazyApURLQuery } from "../../../../lib/query/admin/debug";
import { TextInput } from "../../../../components/form/inputs";
import MutationButton from "../../../../components/form/mutation-button";
import { ApURLResponse } from "../../../../lib/types/debug";
import Loading from "../../../../components/loading";

// Used for syntax highlighting of json result.
import Prism from "../../../../../frontend/prism";

export default function ApURL() {
	const urlField = useTextInput("url");

	const [apURL, apURLResult] = useLazyApURLQuery();
	function submit(e) {
		e.preventDefault();
		apURL(urlField.value ?? "");
	}
	
	return (
		<div className="admin-debug-apurl">
			<form onSubmit={submit}>
				<div className="form-section-docs">
					<h2>AP URL</h2>
					<p>
						Call the provided URL with a signed request, and return the
						raw ActivityPub response in the <code>response_body</code> field.
						<br/>
						Useful for debugging communication between instances, and
						checking the shape of remote objects and actors.
					</p>
				</div>
				<TextInput
					field={urlField}
					label="URL"
					type="url"
					pattern="(http|https):\/\/.+"
					placeholder="https://example.org/users/someone"
				/>
				<MutationButton
					disabled={!urlField.value}
					label="Send request"
					result={apURLResult}
				/>
			</form>
			<ApURLResult {...apURLResult} />
		</div>
	);
}

interface ApURLResultProps {
	isLoading: boolean;
	isFetching: boolean;
	isSuccess: boolean;
	data?: ApURLResponse,
	isError: boolean;
}

function ApURLResult({
	isLoading,
	isFetching,
	isSuccess,
	data,
	isError,
}: ApURLResultProps) {	
	if (!(isSuccess || isError)) {
		// Hasn't been called yet.
		return null;
	}

	if (isLoading || isFetching) {
		return <Loading />;
	}

	if (!data) {
		return "No data";
	}

	const jsonObj = {
		...data,
		response_body: data.response_body.length > 0 ? JSON.parse(data.response_body) : "",
	};

	const jsonStr = JSON.stringify(jsonObj, null, 2);
	return <Highlighted jsonStr={jsonStr} />;
}

function Highlighted({ jsonStr }: { jsonStr: string }) {
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
				<code ref={ref} className="language-json">
					{jsonStr}
				</code>
			</pre>
		</div>
	);
}
