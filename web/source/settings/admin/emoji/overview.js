/*
	 GoToSocial
	 Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
const {Link} = require("wouter");
const splitFilterN = require("split-filter-n");

const NewEmojiForm = require("./new-emoji");

const query = require("../../lib/query");

const base = "/settings/admin/custom-emoji";

module.exports = function EmojiOverview() {
	const {
		data: emoji = [],
		isLoading,
		error
	} = query.useGetAllEmojiQuery({filter: "domain:local"});

	// split all emoji over an object keyed by the category names (or Unsorted)
	const emojiByCategory = React.useMemo(() => splitFilterN(
		emoji,
		[],
		(entry) => entry.category ?? "Unsorted"
	), [emoji]);

	return (
		<>
			<h1>Custom Emoji</h1>
			{error && 
				<div className="error accent">{error}</div>
			}
			{isLoading
				? "Loading..."
				: <>
					<EmojiList emoji={emoji} emojiByCategory={emojiByCategory}/>
					<NewEmojiForm emoji={emoji} emojiByCategory={emojiByCategory}/>
				</>
			}
		</>
	);
};

function EmojiList({emoji, emojiByCategory}) {
	return (
		<div>
			<h2>Overview</h2>
			<div className="list emoji-list">
				{emoji.length == 0 && "No local emoji yet, add one below"}
				{Object.entries(emojiByCategory).map(([category, entries]) => {
					return <EmojiCategory key={category} category={category} entries={entries}/>;
				})}
			</div>
		</div>
	);
}

function EmojiCategory({category, entries}) {
	return (
		<div className="entry">
			<b>{category}</b>
			<div className="emoji-group">
				{entries.map((e) => {
					return (
						<Link key={e.id} to={`${base}/${e.id}`}>
							{/* <Link key={e.static_url} to={`${base}`}> */}
							<a>
								<img src={e.url} alt={e.shortcode} title={`:${e.shortcode}:`}/>
							</a>
						</Link>
					);
				})}
			</div>
		</div>
	);
}