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
const splitFilterN = require("split-filter-n");
const syncpipe = require('syncpipe');
const { matchSorter } = require("match-sorter");

const query = require("../../lib/query");

const ComboBox = require("../../components/combo-box");

function useEmojiByCategory(emoji) {
	// split all emoji over an object keyed by the category names (or Unsorted)
	return React.useMemo(() => splitFilterN(
		emoji,
		[],
		(entry) => entry.category ?? "Unsorted"
	), [emoji]);
}

function CategorySelect({ field, children }) {
	const { value, setIsNew } = field;

	const {
		data: emoji = [],
		isLoading,
		isSuccess,
		error
	} = query.useListEmojiQuery({ filter: "domain:local" });

	const emojiByCategory = useEmojiByCategory(emoji);

	const categories = React.useMemo(() => new Set(Object.keys(emojiByCategory)), [emojiByCategory]);

	// data used by the ComboBox element to select an emoji category
	const categoryItems = React.useMemo(() => {
		return syncpipe(emojiByCategory, [
			(_) => Object.keys(_),            // just emoji category names
			(_) => matchSorter(_, value, { threshold: matchSorter.rankings.NO_MATCH }),  // sorted by complex algorithm
			(_) => _.map((categoryName) => [  // map to input value, and selectable element with icon
				categoryName,
				<>
					<img src={emojiByCategory[categoryName][0].static_url} aria-hidden="true"></img>
					{categoryName}
				</>
			])
		]);
	}, [emojiByCategory, value]);

	React.useEffect(() => {
		if (value != undefined && isSuccess && value.trim().length > 0) {
			setIsNew(!categories.has(value.trim()));
		}
	}, [categories, value, isSuccess, setIsNew]);

	if (error) { // fall back to plain text input, but this would almost certainly have caused a bigger error message elsewhere
		return (
			<>
				<input type="text" placeholder="e.g., reactions" onChange={(e) => { field.value = e.target.value; }} />;
			</>
		);
	} else if (isLoading) {
		return <input type="text" value="Loading categories..." disabled={true} />;
	}

	return (
		<ComboBox
			field={field}
			items={categoryItems}
			label="Category"
			placeholder="e.g., reactions"
			children={children}
		/>
	);
}

module.exports = {
	useEmojiByCategory,
	CategorySelect
};