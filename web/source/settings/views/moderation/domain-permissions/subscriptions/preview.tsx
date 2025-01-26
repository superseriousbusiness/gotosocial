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

import { useTextInput } from "../../../../lib/form";
import { PageableList } from "../../../../components/pageable-list";
import { useLocation } from "wouter";
import { useGetDomainPermissionSubscriptionsPreviewQuery } from "../../../../lib/query/admin/domain-permissions/subscriptions";
import { DomainPermSub } from "../../../../lib/types/domain-permission";
import { Select } from "../../../../components/form/inputs";
import { DomainPermissionSubscriptionDocsLink, SubscriptionListEntry } from "./common";
import { PermType } from "../../../../lib/types/perm";

export default function DomainPermissionSubscriptionsPreview() {
	return (
		<div className="domain-permission-subscriptions-preview">
			<div className="form-section-docs">
				<h1>Domain Permission Subscriptions Preview</h1>
				<p>
					You can use the form below to view through domain permission subscriptions sorted by priority (high to low).
					<br/>
					This reflects the order in which they will actually be fetched by your instance, with higher-priority subscriptions
					creating permissions first, followed by lower-priority subscriptions.
				</p>
				<DomainPermissionSubscriptionDocsLink />
			</div>
			<DomainPermissionSubscriptionsPreviewForm />
		</div>
	);
}

function DomainPermissionSubscriptionsPreviewForm() {
	const [ location, _setLocation ] = useLocation();

	const permType = useTextInput("permission_type", { defaultValue: "block" });
	const {
		data: permSubs,
		isLoading,
		isFetching,
		isSuccess,
		isError,
		error,
	} = useGetDomainPermissionSubscriptionsPreviewQuery(permType.value as PermType);
	
	// Function to map an item to a list entry.
	function itemToEntry(permSub: DomainPermSub): ReactNode {
		return (
			<SubscriptionListEntry
				key={permSub.id}
				permSub={permSub}
				linkTo={`/subscriptions/${permSub.id}`}
				backLocation={location}
			/>
		);
	}

	return (
		<>
			<form>
				<Select
					field={permType}
					label="Permission type"
					options={
						<>
							<option value="block">Block</option>
							<option value="allow">Allow</option>
						</>
					}
				></Select>
			</form>
			<PageableList
				isLoading={isLoading}
				isFetching={isFetching}
				isSuccess={isSuccess}
				items={permSubs}
				itemToEntry={itemToEntry}
				isError={isError}
				error={error}
				emptyMessage={<b>No {permType.value}list subscriptions found.</b>}
			/>
		</>
	);
}
