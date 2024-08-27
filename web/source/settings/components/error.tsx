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

import { SerializedError } from "@reduxjs/toolkit";
import { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import React, { ReactNode } from "react";

function ErrorFallback({ error, resetErrorBoundary }) {
	return (
		<div className="error">
			<p>
				{"An error occured, please report this on the "}
				<a href="https://github.com/superseriousbusiness/gotosocial/issues">GoToSocial issue tracker</a>
				{" or "}
				<a href="https://matrix.to/#/#gotosocial-help:superseriousbusiness.org">Matrix support room</a>.
				<br />Include the details below:
			</p>
			<div className="details">
				<pre>
					{error.name}: {error.message}
				</pre>
				<pre>
					{error.stack}
				</pre>
			</div>
			<p>
				<button onClick={resetErrorBoundary}>Try again</button> or <a href="">refresh the page</a>
			</p>
		</div>
	);
}

interface GtsError {
	/**
	 * Error message returned from the API.
	 */
	error: string;

	/**
	 * For OAuth errors: description of the error.
	 */
	error_description?: string;
}

interface ErrorProps {
	error: FetchBaseQueryError | SerializedError | Error | undefined;
	
	/**
	 * Optional function to clear the error.
	 * If provided, rendered error will have
	 * a "dismiss" button.
	 */
	reset?: () => void;
}

function Error({ error, reset }: ErrorProps) {
	if (error === undefined) {
		return null;
	}
	
	/* eslint-disable-next-line no-console */
	console.error("caught error: ", error);
	
	let message: ReactNode;
	if ("status" in error) {
		if (typeof error.status === "number") {
			// Error containing GTS API error data.
			const gtsError = error.data as GtsError;
			const errMsg = gtsError.error_description ?? gtsError.error;
			message = <>Code {error.status}: {errMsg}</>;
		} else {
			// RTK Query fetching / parsing / timeout error.
			message = <>{error.status}: {error.error}</>;
		}
	} else {
		// SerializedError or Error.
		const errMsg = error.message ?? JSON.stringify(error);
		message = (
			<>{error.name && `${error.name}: `}{errMsg}</>
		);
	}

	let className = "error";
	if (reset) {
		className += " with-dismiss";
	}

	return (
		<div className={className}>
			<span>{message}</span>
			{ reset && 
				<span 
					className="dismiss"
					onClick={reset}
					role="button"
					tabIndex={0}
				>
					<span>Dismiss</span>
					<i className="fa fa-fw fa-close" aria-hidden="true" />
				</span>
			}
		</div>
	);
}

export { ErrorFallback, Error };
