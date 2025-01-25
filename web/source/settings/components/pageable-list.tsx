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

import React, { ReactNode } from "react";
import { useLocation } from "wouter";
import { Error } from "./error";
import { SerializedError } from "@reduxjs/toolkit";
import { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import { Links } from "parse-link-header";
import Loading from "./loading";

export interface PageableListProps<T> {
	isSuccess: boolean;
	items?: T[];
	itemToEntry: (_item: T) => ReactNode;
	isLoading: boolean;
	isFetching?: boolean;
	isError: boolean;
	error: FetchBaseQueryError | SerializedError | undefined;
	emptyMessage: ReactNode;
	prevNextLinks?: Links | null | undefined;
}

export function PageableList<T>({
	isLoading,
	isFetching,
	isSuccess,
	items,
	itemToEntry,
	isError,
	error,
	emptyMessage,
	prevNextLinks,
}: PageableListProps<T>) {
	const [ location, setLocation ] = useLocation();
	
	if (!(isSuccess || isError)) {
		// Hasn't been called yet.
		return null;
	}

	if (isLoading || isFetching) {
		return <Loading />;
	}

	if (error) {
		return <Error error={error} />;
	}

	// Map response to items if possible.
	let content: ReactNode;
	if (items == undefined || items.length == 0) {
		content = <b>{emptyMessage}</b>;
	} else {
		content = (
			<div className="entries">
				{items.map(item => itemToEntry(item))}
			</div>
		);
	}

	// If it's possible to page to next and previous
	// pages, instantiate button handlers for this.
	let prevClick: (() => void) | undefined;
	let nextClick: (() => void) | undefined;
	if (prevNextLinks) {
		const prev = prevNextLinks["prev"];
		if (prev) {
			const prevUrl = new URL(prev.url);
			const prevParams = prevUrl.search;
			prevClick = () => {
				setLocation(location + prevParams.toString());
			};
		}

		const next = prevNextLinks["next"];
		if (next) {
			const nextUrl = new URL(next.url);
			const nextParams = nextUrl.search;
			nextClick = () => {
				setLocation(location + nextParams.toString());
			};
		}
	}

	return (
		<div className="list pageable-list">
			{ content }
			{ prevNextLinks &&
				<div className="prev-next">
					{ prevClick && <button onClick={prevClick}>Previous page</button> }
					{ nextClick && <button onClick={nextClick}>Next page</button> }
				</div>
			}
		</div>
	);
}
