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

import React, { useMemo, useState } from "react";
import { Link } from "wouter";
import { matchSorter } from "match-sorter";
import NewEmojiForm from "./new-emoji";
import { useTextInput } from "../../../../lib/form";
import { useEmojiByCategory } from "../category-select";
import Loading from "../../../../components/loading";
import { Error } from "../../../../components/error";
import { TextInput } from "../../../../components/form/inputs";
import { useListEmojiQuery } from "../../../../lib/query/admin/custom-emoji";
import { CustomEmoji } from "../../../../lib/types/custom-emoji";

export default function EmojiOverview() {
	const { data: emoji = [], isLoading, isError, error } = useListEmojiQuery({ filter: "domain:local" });

	let content: React.JSX.Element;
	if (isLoading) {
		content = <Loading />;
	} else if (isError) {
		content = <Error error={error} />;
	} else {
		content = (
			<>
				<EmojiList emoji={emoji} />
				<NewEmojiForm />
			</>
		);
	}

	return (
		<>
			<h1>Local Custom Emoji</h1>
			<p>
				To use custom emoji in your toots they have to be 'local' to the instance.
				You can either upload them here directly, or copy from those already
				present on other (known) instances through the <Link to={`/remote`}>Remote Emoji</Link> page.
			</p>
			<p>
				<strong>Be warned!</strong> If you upload more than about 300-400 custom emojis in
				total on your instance, this may lead to rate-limiting issues for users and clients
				if they try to load all the emoji images at once (which is what many clients do).
			</p>
			{content}
		</>
	);
}

interface EmojiListParams {
	emoji: CustomEmoji[];
}

function EmojiList({ emoji }: EmojiListParams) {
	const filterField = useTextInput("filter");
	const filter = filterField.value ?? "";
	const emojiByCategory = useEmojiByCategory(emoji);

	// Filter emoji based on shortcode match
	// with user input, hiding empty categories.
	const { filteredEmojis, filteredCount } = useMemo(() => {
		// Amount of emojis removed by the filter.
		// Start with the length of the array since
		// that's the max that can be filtered out.
		let filteredCount = emoji.length;
		
		// Results of the filtering.
		const filteredEmojis: [string, CustomEmoji[]][] = [];
		
		// Filter from emojis in this category.
		emojiByCategory.forEach((entries, category) => {
			const filteredEntries = matchSorter(entries, filter, {
				keys: ["shortcode"]
			});

			if (filteredEntries.length == 0) {
				// Nothing left in this category, don't
				// bother adding it to filteredEmojis.
				return;
			}

			filteredCount -= filteredEntries.length;
			filteredEmojis.push([category, filteredEntries]);
		});

		return { filteredEmojis, filteredCount };
	}, [filter, emojiByCategory, emoji.length]);

	return (
		<>
			<h2>Overview</h2>
			{emoji.length > 0
				? <span>{emoji.length} custom emoji {filteredCount > 0 && `(${filteredCount} filtered)`}</span>
				: <span>No custom emoji yet, you can add one below.</span>
			}
			<div className="list emoji-list">
				<div className="header">
					<TextInput
						field={filterField}
						name="emoji-shortcode"
						placeholder="Search"
						autoCapitalize="none"
						spellCheck="false"
					/>
				</div>
				<div className="entries scrolling">
					{filteredEmojis.length > 0
						? (
							<div className="entries scrolling">
								{filteredEmojis.map(([category, emojis]) => {
									return <EmojiCategory key={category} category={category} emojis={emojis} />;
								})}
							</div>
						)
						: <div className="entry">No local emoji matched your filter.</div>
					}
				</div>
			</div>
		</>
	);
}

interface EmojiCategoryProps {
	category: string;
	emojis: CustomEmoji[];
}

function EmojiCategory({ category, emojis }: EmojiCategoryProps) {
	return (
		<div className="entry">
			<b>{category}</b>
			<div className="emoji-group">
				{emojis.map((emoji) => {
					return (
						<Link key={emoji.id} to={`/local/${emoji.id}`} >
							<EmojiPreview emoji={emoji} />
						</Link>
					);
				})}
			</div>
		</div>
	);
}

function EmojiPreview({ emoji }) {
	const [ animate, setAnimate ] = useState(false);

	return (
		<img
			onMouseEnter={() => { setAnimate(true); }}
			onMouseLeave={() => { setAnimate(false); }}
			src={animate ? emoji.url : emoji.static_url}
			alt={emoji.shortcode}
			title={emoji.shortcode}
			loading="lazy"
		/>
	);
}
