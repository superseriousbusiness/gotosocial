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

import React, { useMemo, useEffect } from "react";
import { useFileInput, useComboBoxInput } from "../../../../lib/form";
import useShortcode from "./use-shortcode";
import useFormSubmit from "../../../../lib/form/submit";
import { TextInput, FileInput } from "../../../../components/form/inputs";
import { CategorySelect } from '../category-select';
import FakeToot from "../../../../components/fake-toot";
import MutationButton from "../../../../components/form/mutation-button";
import { useAddEmojiMutation } from "../../../../lib/query/admin/custom-emoji";
import { useInstanceV1Query } from "../../../../lib/query/gts-api";

export default function NewEmojiForm() {
	const shortcode = useShortcode();

	const { data: instance } = useInstanceV1Query();
	const emojiMaxSize = useMemo(() => {
		return instance?.configuration?.emojis?.emoji_size_limit ?? 50 * 1024;
	}, [instance]);

	const image = useFileInput("image", {
		withPreview: true,
		maxSize: emojiMaxSize
	});

	const category = useComboBoxInput("category");

	const [submitForm, result] = useFormSubmit({
		shortcode, image, category
	}, useAddEmojiMutation());

	useEffect(() => {
		if (shortcode.value === undefined || shortcode.value.length == 0) {
			if (image.value != undefined) {
				let [name, _ext] = image.value.name.split(".");
				shortcode.setter(name);
			}
		}

		/* We explicitly don't want to have 'shortcode' as a dependency here
			 because we only want to change the shortcode to the filename if the field is empty
			 at the moment the file is selected, not some time after when the field is emptied
		*/
		/* eslint-disable-next-line react-hooks/exhaustive-deps */
	}, [image.value]);

	let emojiOrShortcode;

	if (image.previewValue != undefined) {
		emojiOrShortcode = <img
			className="emoji"
			src={image.previewValue}
			title={`:${shortcode.value}:`}
			alt={shortcode.value}
		/>;
	} else if (shortcode.value !== undefined && shortcode.value.length > 0) {
		emojiOrShortcode = `:${shortcode.value}:`;
	} else {
		emojiOrShortcode = `:your_emoji_here:`;
	}

	return (
		<div>
			<h2>Add new custom emoji</h2>

			<FakeToot>
				Look at this new custom emoji {emojiOrShortcode} isn&apos;t it cool?
			</FakeToot>

			<form onSubmit={submitForm} className="form-flex">
				<FileInput
					field={image}
					accept="image/png,image/gif,image/webp"
				/>

				<TextInput
					field={shortcode}
					label="Shortcode, must be unique among the instance's local emoji"
				/>

				<CategorySelect
					field={category}
					children={[]}
				/>

				<MutationButton
					disabled={image.previewValue === undefined}
					label="Upload emoji"
					result={result}
				/>
			</form>
		</div>
	);
}
