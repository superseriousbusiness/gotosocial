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

const query = require("../../../lib/query");

const {
	useFileInput,
	useComboBoxInput
} = require("../../../lib/form");
const useShortcode = require("./use-shortcode");

const useFormSubmit = require("../../../lib/form/submit");

const {
	TextInput, FileInput
} = require("../../../components/form/inputs");

const { CategorySelect } = require('../category-select');
const FakeToot = require("../../../components/fake-toot");
const MutationButton = require("../../../components/form/mutation-button");

module.exports = function NewEmojiForm() {
	const shortcode = useShortcode();

	const image = useFileInput("image", {
		withPreview: true,
		maxSize: 50 * 1024 // TODO: get from instance api?
	});

	const category = useComboBoxInput("category");

	const [submitForm, result] = useFormSubmit({
		shortcode, image, category
	}, query.useAddEmojiMutation());

	React.useEffect(() => {
		if (shortcode.value.length == 0) {
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

	let emojiOrShortcode = `:${shortcode.value}:`;

	if (image.previewValue != undefined) {
		emojiOrShortcode = <img
			className="emoji"
			src={image.previewValue}
			title={`:${shortcode.value}:`}
			alt={shortcode.value}
		/>;
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
					accept="image/png,image/gif"
				/>

				<TextInput
					field={shortcode}
					label="Shortcode, must be unique among the instance's local emoji"
				/>

				<CategorySelect
					field={category}
				/>

				<MutationButton label="Upload emoji" result={result} />
			</form>
		</div>
	);
};