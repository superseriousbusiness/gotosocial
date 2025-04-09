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

import React, { useEffect } from "react";
import { useLocation } from "wouter";
import { AdminAccount } from "../lib/types/account";
import { useLazyGetAccountQuery } from "../lib/query/admin";
import Loading from "./loading";
import { Error as ErrorC } from "./error";

interface UsernameLozengeProps {
	/**
	 * Either an account ID (for fetching) or an account.
	 */
	account?: string | AdminAccount;
	/**
	 * Make the lozenge clickable and link to this location.
	 */
	linkTo?: string;
	/**
	 * Location to set as backLocation after linking to linkTo.
	 */
	backLocation?: string;
	/**
	 * Additional classnames to add to the lozenge.
	 */
	classNames?: string[];
}

export default function UsernameLozenge({ account, linkTo, backLocation, classNames }: UsernameLozengeProps) {
	if (account === undefined) {
		return <>[unknown]</>;
	} else if (typeof account === "string") {
		return (
			<FetchUsernameLozenge
				accountID={account}
				linkTo={linkTo}
				backLocation={backLocation}
				classNames={classNames}
			/>
		);
	} else {
		return (
			<ReadyUsernameLozenge
				account={account}
				linkTo={linkTo}
				backLocation={backLocation}
				classNames={classNames}
			/>
		);
	}

}

interface FetchUsernameLozengeProps {
	accountID: string;
	linkTo?: string;
	backLocation?: string;
	classNames?: string[];
}

function FetchUsernameLozenge({ accountID, linkTo, backLocation, classNames }: FetchUsernameLozengeProps) {
	const [ trigger, result ] = useLazyGetAccountQuery();
	
	// Call to get the account
	// using the provided ID.
	useEffect(() => {
		trigger(accountID, true);
	}, [trigger, accountID]);

	const {
		data: account,
		isLoading,
		isFetching,
		isError,
		error,
	} = result;

	// Wait for the account
	// model to be returned.
	if (isError) {
		return <ErrorC error={error} />;
	} else if (isLoading || isFetching || account === undefined) {
		return <Loading />;
	}

	return (
		<ReadyUsernameLozenge
			account={account}
			linkTo={linkTo}
			backLocation={backLocation}
			classNames={classNames}
		/>
	);
}

interface ReadyUsernameLozengeProps {
	account: AdminAccount;
	linkTo?: string;
	backLocation?: string;
	classNames?: string[];
}

function ReadyUsernameLozenge({ account, linkTo, backLocation, classNames }: ReadyUsernameLozengeProps) {
	const [ _location, setLocation ] = useLocation();
	
	let className = "username-lozenge";
	let isLocal = account.domain == null;

	if (account.suspended) {
		className += " suspended";
	}

	if (isLocal) {
		className += " local";
	}

	if (classNames) {
		className = [ className, classNames ].flat().join(" ");
	}

	let icon = isLocal
		? { fa: "fa-home", info: "Local user" }
		: { fa: "fa-external-link-square", info: "Remote user" };

	const content = (
		<>
			<i className={`fa fa-fw ${icon.fa}`} aria-hidden="true" title={icon.info} />
			<span className="sr-only">{icon.info}</span>
			&nbsp;
			<span className="acct">@{account.account.acct}</span>
		</>
	);

	if (linkTo) {
		className += " pseudolink";
		const onClick = () => {
			// When clicking on an account, direct
			// to the detail view for that account.
			setLocation(linkTo, {
				// Store the back location in history so
				// the detail view can use it to return to
				// this page (including query parameters).
				state: { backLocation: backLocation }
			});
		};
		return (
			<span
				className={className}
				onClick={onClick}
				onKeyDown={e => e.key === "Enter" && onClick()}
				role="link"
				tabIndex={0}
			>
				{content}
			</span>
		);
	} else {
		return (
			<div className={className}>
				{content}
			</div>
		);
	}
}
