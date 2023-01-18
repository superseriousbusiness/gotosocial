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
const { useRoute, Link, Redirect } = require("wouter");

const query = require("../../../lib/query");

const { useComboBoxInput, useFileInput, useValue } = require("../../../lib/form");
const { CategorySelect } = require("../category-select");

const useFormSubmit = require("../../../lib/form/submit");

const FakeToot = require("../../../components/fake-toot");
const FormWithData = require("../../../lib/form/form-with-data");
const Loading = require("../../../components/loading");
const { FileInput } = require("../../../components/form/inputs");
const MutationButton = require("../../../components/form/mutation-button");
const { Error } = require("../../../components/error");

const base = "/settings/custom-emoji/local";

module.exports = function EmojiDetailRoute() {
	let [_match, params] = useRoute(`${base}/:emojiId`);
	if (params?.emojiId == undefined) {
		return <Redirect to={base} />;
	} else {
		return (
			<div className="emoji-detail">
				<Link to={base}><a>&lt; go back</a></Link>
				<FormWithData dataQuery={query.useGetEmojiQuery} queryArg={params.emojiId} DataForm={EmojiDetailForm} />
			</div>
		);
	}
};

function EmojiDetailForm({ data: emoji }) {
	const form = {
		id: useValue("id", emoji.id),
		category: useComboBoxInput("category", { defaultValue: emoji.category }),
		image: useFileInput("image", {
			withPreview: true,
			maxSize: 50 * 1024 // TODO: get from instance api
		})
	};

	const [modifyEmoji, result] = useFormSubmit(form, query.useEditEmojiMutation());

	// Automatic submitting of category change
	React.useEffect(() => {
		if (
			form.category.hasChanged() &&
			!form.category.state.open &&
			!form.category.isNew) {
			modifyEmoji();
		}
		/* eslint-disable-next-line react-hooks/exhaustive-deps */
	}, [form.category.hasChanged(), form.category.isNew, form.category.state.open]);

	const [deleteEmoji, deleteResult] = query.useDeleteEmojiMutation();

	if (deleteResult.isSuccess) {
		return <Redirect to={base} />;
	}

	return (
		<>
			<div className="emoji-header">
				<img src={emoji.url} alt={emoji.shortcode} title={emoji.shortcode} />
				<div>
					<h2>{emoji.shortcode}</h2>
					<MutationButton
						label="Delete"
						type="button"
						onClick={() => deleteEmoji(emoji.id)}
						className="danger"
						showError={false}
						result={deleteResult}
					/>
				</div>
			</div>

			<form onSubmit={modifyEmoji} className="left-border">
				<h2>Modify this emoji {result.isLoading && <Loading />}</h2>

				<div className="update-category">
					<CategorySelect
						field={form.category}
					>
						<MutationButton
							name="create-category"
							label="Create"
							result={result}
							showError={false}
							style={{ visibility: (form.category.isNew ? "initial" : "hidden") }}
						/>
					</CategorySelect>
				</div>

				<div className="update-image">
					<FileInput
						field={form.image}
						label="Image"
						accept="image/png,image/gif"
					/>

					<MutationButton
						name="image"
						label="Replace image"
						showError={false}
						result={result}
					/>

					<FakeToot>
						Look at this new custom emoji <img
							className="emoji"
							src={form.image.previewURL ?? emoji.url}
							title={`:${emoji.shortcode}:`}
							alt={emoji.shortcode}
						/> isn&apos;t it cool?
					</FakeToot>

					{result.error && <Error error={result.error} />}
					{deleteResult.error && <Error error={deleteResult.error} />}
				</div>
			</form>
		</>
	);
}