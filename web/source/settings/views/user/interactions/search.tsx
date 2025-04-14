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

import { useBoolInput, useTextInput } from "../../../lib/form";
import { PageableList } from "../../../components/pageable-list";
import MutationButton from "../../../components/form/mutation-button";
import { useLocation, useSearch } from "wouter";
import { useApproveInteractionRequestMutation, useLazySearchInteractionRequestsQuery, useRejectInteractionRequestMutation } from "../../../lib/query/user/interactions";
import { InteractionRequest } from "../../../lib/types/interaction";
import { Checkbox } from "../../../components/form/inputs";
import { useContent, useIcon, useNoun, useVerbed } from "./util";

function defaultTrue(urlQueryVal: string | null): boolean {
	if (urlQueryVal === null) {
		return true;
	}
	
	return urlQueryVal.toLowerCase() !== "false";
}

export default function InteractionRequestsSearchForm() {
	const [ location, setLocation ] = useLocation();
	const search = useSearch();
	const urlQueryParams = useMemo(() => new URLSearchParams(search), [search]);
	const [ searchReqs, searchRes ] = useLazySearchInteractionRequestsQuery();

	// Populate search form using values from
	// urlQueryParams, to allow paging.
	const form = {
		statusID: useTextInput("status_id", {
			defaultValue: urlQueryParams.get("status_id") ?? ""
		}),
		likes: useBoolInput("favourites", {
			defaultValue: defaultTrue(urlQueryParams.get("favourites"))
		}),
		replies: useBoolInput("replies", {
			defaultValue: defaultTrue(urlQueryParams.get("replies"))
		}),
		boosts: useBoolInput("reblogs", {
			defaultValue: defaultTrue(urlQueryParams.get("reblogs"))
		}),
	};

	// On mount, trigger search.
	useEffect(() => {
		searchReqs(Object.fromEntries(urlQueryParams), true);
	}, [urlQueryParams, searchReqs]);

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
	// "back" on the interaction req detail view.
	const backLocation = location + (urlQueryParams.size > 0 ? `?${urlQueryParams}` : "");
	
	// Function to map an item to a list entry.
	function itemToEntry(req: InteractionRequest): ReactNode {
		return (
			<ReqsListEntry
				key={req.id}
				req={req}
				linkTo={`/${req.id}`}
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
				<Checkbox
					label="Include likes"
					field={form.likes}
				/>
				<Checkbox
					label="Include replies"
					field={form.replies}
				/>
				<Checkbox
					label="Include boosts"
					field={form.boosts}
				/>
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
				items={searchRes.data?.requests}
				itemToEntry={itemToEntry}
				isError={searchRes.isError}
				error={searchRes.error}
				emptyMessage={<b>No interaction requests found that match your query.</b>}
				prevNextLinks={searchRes.data?.links}
			/>
		</>
	);
}

interface ReqsListEntryProps {
	req: InteractionRequest;
	linkTo: string;
	backLocation: string;
}

function ReqsListEntry({ req, linkTo, backLocation }: ReqsListEntryProps) {
	const [ _location, setLocation ] = useLocation();
	
	const [ approve, approveResult ] = useApproveInteractionRequestMutation();
	const [ reject, rejectResult ] = useRejectInteractionRequestMutation();
	
	const verbed = useVerbed(req.type);
	const noun = useNoun(req.type);
	const icon = useIcon(req.type);
	
	const strap = useMemo(() => {
		return "@" + req.account.acct + " " + verbed + " your post.";
	}, [req.account, verbed]);
	
	const label = useMemo(() => {
		return noun + " from @" + req.account.acct;
	}, [req.account, noun]);

	const ourContent = useContent(req.status);
	const theirContent = useContent(req.reply);

	const onClick = (e) => {
		e.preventDefault();
		// When clicking on a request, direct
		// to the detail view for that request.
		setLocation(linkTo, {
			// Store the back location in history so
			// the detail view can use it to return to
			// this page (including query parameters).
			state: { backLocation: backLocation }
		});
	};

	return (
		<span
			className={`pseudolink entry interaction-request`}
			aria-label={label}
			title={label}
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
			<span className="text-cutoff">
				<i
					className={`fa fa-fw ${icon}`}
					aria-hidden="true"
				/> <strong>{strap}</strong>
			</span>
			<dl className="info-list">
				<div className="info-list-entry">
					<dt>You wrote:</dt>
					<dd className="text-cutoff">
						{ourContent}
					</dd>
				</div>
				{ req.type === "reply" &&
					<div className="info-list-entry">
						<dt>They wrote:</dt>
						<dd className="text-cutoff">
							{theirContent}
						</dd>
					</div>
				}
			</dl>
			<div className="action-buttons">
				<MutationButton
					label="Accept"
					title={`Accept ${noun}`}
					type="button"
					className="button"
					onClick={(e) => {
						e.preventDefault();
						e.stopPropagation();
						approve(req.id);
					}}
					disabled={false}
					showError={false}
					result={approveResult}
				/>

				<MutationButton
					label="Reject"
					title={`Reject ${noun}`}
					type="button"
					className="button danger"
					onClick={(e) => {
						e.preventDefault();
						e.stopPropagation();
						reject(req.id);
					}}
					disabled={false}
					showError={false}
					result={rejectResult}
				/>
			</div>
		</span>
	);
}

