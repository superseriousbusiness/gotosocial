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

import React, { useEffect } from "react";
import { useRoute, Link, Redirect } from "wouter";

import { useComboBoxInput, useFileInput, useValue } from "../../../lib/form";
import { CategorySelect } from "../category-select";

import useFormSubmit from "../../../lib/form/submit";
import { useBaseUrl } from "../../../lib/navigation/util";

import FakeToot from "../../../components/fake-toot";
import FormWithData from "../../../lib/form/form-with-data";
import Loading from "../../../components/loading";
import { FileInput } from "../../../components/form/inputs";
import MutationButton from "../../../components/form/mutation-button";
import { Error } from "../../../components/error";

import { useGetEmojiQuery, useEditEmojiMutation, useDeleteEmojiMutation } from "../../../lib/query/admin/custom-emoji";

export default function EmojiDetailRoute({ }) {
	const baseUrl = useBaseUrl();
	let [_match, params] = useRoute(`${baseUrl}/:emojiId`);
	if (params?.emojiId == undefined) {
		return <Redirect to={baseUrl} />;
	} else {
		return (
			<div className="emoji-detail">
				<Link to={baseUrl}><a>&lt; go back</a></Link>
				<FormWithData dataQuery={useGetEmojiQuery} queryArg={params.emojiId} DataForm={EmojiDetailForm} />
			</div>
		);
	}
};

function EmojiDetailForm({ data: emoji }) {
	const baseUrl = useBaseUrl();
	const form = {
		id: useValue("id", emoji.id),
		category: useComboBoxInput("category", { source: emoji }),
		image: useFileInput("image", {
			withPreview: true,
			maxSize: 50 * 1024 // TODO: get from instance api
		})
	};

	const [modifyEmoji, result] = useFormSubmit(form, useEditEmojiMutation());

	// Automatic submitting of category change
	useEffect(() => {
		if (
			form.category.hasChanged() &&
			!form.category.state.open &&
			!form.category.isNew) {
			modifyEmoji();
		}
		/* eslint-disable-next-line react-hooks/exhaustive-deps */
	}, [form.category.hasChanged(), form.category.isNew, form.category.state.open]);

	const [deleteEmoji, deleteResult] = useDeleteEmojiMutation();

	if (deleteResult.isSuccess) {
		return <Redirect to={baseUrl} />;
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