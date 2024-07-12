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

import React, { useMemo, useEffect, ReactNode } from "react";
import { useFileInput, useComboBoxInput } from "../../../../lib/form";
import useShortcode from "./use-shortcode";
import useFormSubmit from "../../../../lib/form/submit";
import { TextInput, FileInput } from "../../../../components/form/inputs";
import { CategorySelect } from '../category-select';
import { FakeStatus } from "../../../../components/status";
import MutationButton from "../../../../components/form/mutation-button";
import { useAddEmojiMutation } from "../../../../lib/query/admin/custom-emoji";
import { useInstanceV1Query } from "../../../../lib/query/gts-api";
import prettierBytes from "prettier-bytes";

export default function NewEmojiForm() {
	const { data: instance } = useInstanceV1Query();
	const emojiMaxSize = useMemo(() => {
		return instance?.configuration?.emojis?.emoji_size_limit ?? 50 * 1024;
	}, [instance]);

	const prettierMaxSize = useMemo(() => {
		return prettierBytes(emojiMaxSize);
	}, [emojiMaxSize]);

	const form = {
		shortcode: useShortcode(),
		image: useFileInput("image", {
			withPreview: true,
			maxSize: emojiMaxSize
		}),
		category: useComboBoxInput("category"),
	};

	const [submitForm, result] = useFormSubmit(
		form,
		useAddEmojiMutation(),
		{
			changedOnly: false,
			// On submission, reset form values
			// no matter what the result was.
			onFinish: (_res) => {
				form.shortcode.reset();
				form.image.reset();
				form.category.reset();
			}
		},
	);

	useEffect(() => {
		// If shortcode has not been entered yet, but an image file
		// has been submitted, suggest a shortcode based on filename.
		if (
			(form.shortcode.value === undefined || form.shortcode.value.length === 0) &&
			form.image.value !== undefined
		) {
			let [name, _ext] = form.image.value.name.split(".");
			form.shortcode.setter(name);
		}

		// We explicitly don't want to have 'shortcode' as a
		// dependency here because we only want to change the
		// shortcode to the filename if the field is empty at
		// the moment the file is selected, not some time after
		// when the field is emptied.
		//
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [form.image.value]);

	let emojiOrShortcode: ReactNode;
	if (form.image.previewValue !== undefined) {
		emojiOrShortcode = (
			<img
				className="emoji"
				src={form.image.previewValue}
				title={`:${form.shortcode.value}:`}
				alt={form.shortcode.value}
			/>
		);
	} else if (form.shortcode.value !== undefined && form.shortcode.value.length > 0) {
		emojiOrShortcode = `:${form.shortcode.value}:`;
	} else {
		emojiOrShortcode = `:your_emoji_here:`;
	}

	return (
		<div>
			<h2>Add new custom emoji</h2>

			<FakeStatus>
				Look at this new custom emoji {emojiOrShortcode} isn&apos;t it cool?
			</FakeStatus>

			<form onSubmit={submitForm} className="form-flex">
				<FileInput
					field={form.image}
					label={`Image file: png, gif, or static webp; max size ${prettierMaxSize}`}
					accept="image/png,image/gif,image/webp"
				/>

				<TextInput
					field={form.shortcode}
					label="Shortcode, must be unique among the instance's local emoji"
					autoCapitalize="none"
					spellCheck="false"
					{...{pattern: "^\\w{2,30}$"}}
				/>

				<CategorySelect
					field={form.category}
				/>

				<MutationButton
					disabled={form.image.previewValue === undefined || form.shortcode.value?.length === 0}
					label="Upload emoji"
					result={result}
				/>
			</form>
		</div>
	);
}
