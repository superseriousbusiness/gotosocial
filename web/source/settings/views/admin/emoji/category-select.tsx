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

import React, { useMemo, useEffect, PropsWithChildren, ReactElement } from "react";
import { matchSorter } from "match-sorter";
import ComboBox from "../../../components/combo-box";
import { useListEmojiQuery } from "../../../lib/query/admin/custom-emoji";
import { CustomEmoji } from "../../../lib/types/custom-emoji";
import { ComboboxFormInputHook } from "../../../lib/form/types";
import Loading from "../../../components/loading";
import { Error } from "../../../components/error";

/**
 * Sort all emoji into a map keyed by
 * the category names (or "Unsorted").
 */
export function useEmojiByCategory(emojis: CustomEmoji[]) {	
	return useMemo(() => {
		const byCategory = new Map<string, CustomEmoji[]>();
		
		emojis.forEach((emoji) => {
			const key = emoji.category ?? "Unsorted";
			const value = byCategory.get(key) ?? [];
			value.push(emoji);
			byCategory.set(key, value);
		});

		return byCategory;
	}, [emojis]);
}

interface CategorySelectProps {
	field: ComboboxFormInputHook;
}

/**
 * 
 * Renders a cute lil searchable "category select" dropdown.
 */
export function CategorySelect({ field, children }: PropsWithChildren<CategorySelectProps>) {
	// Get all local emojis.
	const {
		data: emoji = [],
		isLoading,
		isSuccess,
		isError,
		error,
	} = useListEmojiQuery({ filter: "domain:local" });

	const emojiByCategory = useEmojiByCategory(emoji);	
	const categories = useMemo(() => new Set(emojiByCategory.keys()), [emojiByCategory]);
	const { value, setIsNew } = field;

	// Data used by the ComboBox element
	// to select an emoji category.
	const categoryItems = useMemo(() => {
		const categoriesArr = Array.from(categories); 

		// Sorted by complex algorithm.
		const categoryNames = matchSorter(
			categoriesArr,
			value ?? "",
			{ threshold: matchSorter.rankings.NO_MATCH },
		);

		// Map each category to the static image
		// of the first emoji it contains.
		const categoryItems: [string, ReactElement][] = [];
		categoryNames.forEach((categoryName) => {
			let src: string | undefined;
			const items = emojiByCategory.get(categoryName);
			if (items && items.length > 0) {
				src = items[0].static_url;
			}

			categoryItems.push([
				categoryName,
				<>
					<img
						src={src}
						aria-hidden="true"
					/>
					{categoryName}
				</>
			]);
		});

		return categoryItems;
	}, [emojiByCategory, categories, value]);

	// New category if something has been entered
	// and we don't have it in categories yet.
	useEffect(() => {
		if (value !== undefined) {
			const trimmed = value.trim();
			if (trimmed.length > 0) {
				setIsNew(!categories.has(trimmed));
			}
		}
	}, [categories, value, isSuccess, setIsNew]);

	if (isLoading) {
		return <Loading />;
	} else if (isError) {
		return <Error error={error} />;
	} else {
		return (
			<ComboBox
				field={field}
				items={categoryItems}
				label="Category"
				placeholder="e.g., reactions"
				autoCapitalize="none"
				spellCheck="false"
			>
				{children}
			</ComboBox>
		);
	}
}
