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
const { Link } = require("wouter");
const syncpipe = require("syncpipe");
const { matchSorter } = require("match-sorter");

const NewEmojiForm = require("./new-emoji");
const { useTextInput } = require("../../../lib/form");

const query = require("../../../lib/query");
const { useEmojiByCategory } = require("../category-select");

const Loading = require("../../../components/loading");
const { Error } = require("../../../components/error");
const { TextInput } = require("../../../components/form/inputs");

module.exports = function EmojiOverview({ baseUrl }) {
	const {
		data: emoji = [],
		isLoading,
		isError,
		error
	} = query.useListEmojiQuery({ filter: "domain:local" });

	let content = null;

	if (isLoading) {
		content = <Loading />;
	} else if (isError) {
		content = <Error error={error} />;
	} else {
		content = (
			<>
				<EmojiList emoji={emoji} baseUrl={baseUrl} />
				<NewEmojiForm emoji={emoji} />
			</>
		);
	}

	return (
		<>
			<h1>Local Custom Emoji</h1>
			<p>
				To use custom emoji in your toots they have to be 'local' to the instance.
				You can either upload them here directly, or copy from those already
				present on other (known) instances through the <Link to={`../remote`}>Remote Emoji</Link> page.
			</p>
			{content}
		</>
	);
};

function EmojiList({ emoji, baseUrl }) {
	const filterField = useTextInput("filter");
	const filter = filterField.value;

	const emojiByCategory = useEmojiByCategory(emoji);

	/* Filter emoji based on shortcode match with user input, hiding empty categories */
	const { filteredEmoji, hidden } = React.useMemo(() => {
		let hidden = emoji.length;
		const filteredEmoji = syncpipe(emojiByCategory, [
			(_) => Object.entries(emojiByCategory),
			(_) => _.map(([category, entries]) => {
				let filteredEntries = matchSorter(entries, filter, { keys: ["shortcode"] });
				if (filteredEntries.length == 0) {
					return null;
				} else {
					hidden -= filteredEntries.length;
					return [category, filteredEntries];
				}
			}),
			(_) => _.filter((value) => value !== null)
		]);

		return { filteredEmoji, hidden };
	}, [filter, emojiByCategory, emoji.length]);

	return (
		<div>
			<h2>Overview</h2>
			{emoji.length > 0
				? <span>{emoji.length} custom emoji {hidden > 0 && `(${hidden} filtered)`}</span>
				: <span>No custom emoji yet, you can add one below.</span>
			}
			<div className="list emoji-list">
				<div className="header">
					<TextInput
						field={filterField}
						name="emoji-shortcode"
						placeholder="Search"
					/>
				</div>
				<div className="entries scrolling">
					{filteredEmoji.length > 0
						? (
							<div className="entries scrolling">
								{filteredEmoji.map(([category, entries]) => {
									return <EmojiCategory key={category} category={category} entries={entries} baseUrl={baseUrl} />;
								})}
							</div>
						)
						: <div className="entry">No local emoji matched your filter.</div>
					}
				</div>
			</div>
		</div>
	);
}

function EmojiCategory({ category, entries, baseUrl }) {
	return (
		<div className="entry">
			<b>{category}</b>
			<div className="emoji-group">
				{entries.map((e) => {
					return (
						<Link key={e.id} to={`${baseUrl}/${e.id}`}>
							<a>
								<img src={e.url} alt={e.shortcode} title={`:${e.shortcode}:`} />
							</a>
						</Link>
					);
				})}
			</div>
		</div>
	);
}