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

import React, { ReactNode, useEffect, useMemo } from "react";

import { useTextInput } from "../../../lib/form";
import { PageableList } from "../../../components/pageable-list";
import MutationButton from "../../../components/form/mutation-button";
import { useLocation, useSearch } from "wouter";
import { Select } from "../../../components/form/inputs";
import { useLazySearchAppQuery } from "../../../lib/query/user/applications";
import { App } from "../../../lib/types/application";
import { useAppWebsite, useCreated, useRedirectURIs } from "./common";

export default function ApplicationsSearchForm() {
	const [ location, setLocation ] = useLocation();
	const search = useSearch();
	const urlQueryParams = useMemo(() => new URLSearchParams(search), [search]);
	const [ searchApps, searchRes ] = useLazySearchAppQuery();

	// Populate search form using values from
	// urlQueryParams, to allow paging.
	const form = {
		limit: useTextInput("limit", { defaultValue: urlQueryParams.get("limit") ?? "20" })
	};

	// On mount, trigger search.
	useEffect(() => {
		searchApps(Object.fromEntries(urlQueryParams), true);
	}, [urlQueryParams, searchApps]);

	// Rather than triggering the search directly,
	// the "submit" button changes the location
	// based on form field params, and lets the
	// useEffect hook above actually do the search.
	function submitQuery(e) {
		e.preventDefault();

		// Parse query parameters.
		const entries = Object.entries(form).map(([k, v]) => {
			// Take only defined form fields.
			if (v.value === undefined) {
				return null;
			} else if (typeof v.value === "string" && v.value.length === 0) {
				return null;
			}

			return [[k, v.value.toString()]];
		}).flatMap(kv => {
			// Remove any nulls.
			return kv !== null ? kv : [];
		});

		const searchParams = new URLSearchParams(entries);
		setLocation(location + "?" + searchParams.toString());
	}

	// Location to return to when user clicks
	// "back" on the application detail view.
	const backLocation = location + (urlQueryParams.size > 0 ? `?${urlQueryParams}` : "");

	// Function to map an item to a list entry.
	function itemToEntry(application: App): ReactNode {
		return (
			<ApplicationListEntry
				key={application.id}
				app={application}
				linkTo={`/${application.id}`}
				backLocation={backLocation}
			/>
		);
	}

	return (
		<>
			<form
				onSubmit={submitQuery}
				// Prevent password managers
				// trying to fill in fields.
				autoComplete="off"
			>
				<Select
					field={form.limit}
					label="Items per page"
					options={
						<>
							<option value="20">20</option>
							<option value="50">50</option>
							<option value="0">No limit / show all</option>
						</>
					}
				></Select>
				<MutationButton
					disabled={false}
					label={"Search"}
					result={searchRes}
				/>
			</form>
			<PageableList
				isLoading={searchRes.isLoading}
				isFetching={searchRes.isFetching}
				isSuccess={searchRes.isSuccess}
				items={searchRes.data?.apps}
				itemToEntry={itemToEntry}
				isError={searchRes.isError}
				error={searchRes.error}
				emptyMessage={<b>No applications found.</b>}
				prevNextLinks={searchRes.data?.links}
			/>
		</>
	);
}

interface ApplicationListEntryProps {
	app: App;
	linkTo: string;
	backLocation: string;
}

function ApplicationListEntry({ app, linkTo, backLocation }: ApplicationListEntryProps) {
	const [ _location, setLocation ] = useLocation();
	const appWebsite = useAppWebsite(app);
	const created = useCreated(app);
	const redirectURIs = useRedirectURIs(app);

	const onClick = (e) => {
		e.preventDefault();
		// When clicking on an app, direct
		// to the detail view for that app.
		setLocation(linkTo, {
			// Store the back location in history so
			// the detail view can use it to return to
			// this page (including query parameters).
			state: { backLocation: backLocation }
		});
	};

	return (
		<span
			className={`pseudolink application entry`}
			aria-label={`${app.name}`}
			title={`${app.name}`}
			onClick={onClick}
			onKeyDown={(e) => {
				if (e.key === "Enter") {
					e.preventDefault();
					onClick(e);
				}
			}}
			role="link"
			tabIndex={0}
		>
			<dl className="info-list">
				<div className="info-list-entry">
					<dt>Name:</dt>
					<dd className="text-cutoff">{app.name}</dd>
				</div>

				{ appWebsite && 
					<div className="info-list-entry">
						<dt>Website:</dt>
						<dd className="text-cutoff">{appWebsite}</dd>
					</div>
				}

				<div className="info-list-entry">
					<dt>Created:</dt>
					<dd className="text-cutoff">{created}</dd>
				</div>

				<div className="info-list-entry">
					<dt>Scopes:</dt>
					<dd className="text-cutoff monospace">{app.scopes.join(" ")}</dd>
				</div>

				<div className="info-list-entry">
					<dt>Redirect URI(s):</dt>
					<dd className="text-cutoff monospace">{redirectURIs}</dd>
				</div>
			</dl>
		</span>
	);
}
