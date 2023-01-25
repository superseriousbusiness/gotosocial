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
const { Link, useLocation } = require("wouter");
const { matchSorter } = require("match-sorter");

const { useTextInput } = require("../../lib/form");

const { TextInput } = require("../../components/form/inputs");

const query = require("../../lib/query");

const Loading = require("../../components/loading");

module.exports = function InstanceOverview({ baseUrl }) {
	const { data: blockedInstances = [], isLoading } = query.useInstanceBlocksQuery();

	const [_location, setLocation] = useLocation();

	const filterField = useTextInput("filter");
	const filter = filterField.value;

	const blockedInstancesList = React.useMemo(() => {
		return Object.values(blockedInstances);
	}, [blockedInstances]);

	const filteredInstances = React.useMemo(() => {
		return matchSorter(blockedInstancesList, filter, { keys: ["domain"] });
	}, [blockedInstancesList, filter]);

	let filtered = blockedInstancesList.length - filteredInstances.length;

	function filterFormSubmit(e) {
		e.preventDefault();
		setLocation(`${baseUrl}/${filter}`);
	}

	if (isLoading) {
		return <Loading />;
	}

	return (
		<>
			<h1>Federation</h1>

			<div className="instance-list">
				<h2>Suspended instances</h2>
				<p>
					Suspending a domain blocks all current and future accounts on that instance. Stored content will be removed,
					and no more data is sent to the remote server.<br />
					This extends to all subdomains as well, so blocking 'example.com' also includes 'social.example.com'.
				</p>
				<form className="filter" role="search" onSubmit={filterFormSubmit}>
					<TextInput field={filterField} placeholder="example.com" label="Search or add domain suspension" />
					<Link to={`${baseUrl}/${filter}`}><a className="button">Suspend</a></Link>
				</form>
				<div>
					<span>
						{blockedInstancesList.length} blocked instance{blockedInstancesList.length != 1 ? "s" : ""} {filtered > 0 && `(${filtered} filtered by search)`}
					</span>
					<div className="list">
						<div className="entries scrolling">
							{filteredInstances.map((entry) => {
								return (
									<Link key={entry.domain} to={`${baseUrl}/${entry.domain}`}>
										<a className="entry nounderline">
											<span id="domain">
												{entry.domain}
											</span>
											<span id="date">
												{new Date(entry.created_at).toLocaleString()}
											</span>
										</a>
									</Link>
								);
							})}
						</div>
					</div>
				</div>
			</div>
			<Link to={`${baseUrl}/import-export`}><a>Or use the bulk import/export interface</a></Link>
		</>
	);
};