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

const { useRoute, Link, Redirect } = require("wouter");

const { CategorySelect } = require("../category-select");
const { useComboBoxInput, useFileInput } = require("../../../components/form");

const query = require("../../../lib/query");
const FakeToot = require("../../../components/fake-toot");
const Loading = require("../../../components/loading");

const base = "/settings/custom-emoji/local";

module.exports = function EmojiDetailRoute() {
	let [_match, params] = useRoute(`${base}/:emojiId`);
	if (params?.emojiId == undefined) {
		return <Redirect to={base}/>;
	} else {
		return (
			<div className="emoji-detail">
				<Link to={base}><a>&lt; go back</a></Link>
				<EmojiDetailData emojiId={params.emojiId}/>
			</div>
		);
	}
};

function EmojiDetailData({emojiId}) {
	const {currentData: emoji, isLoading, error} = query.useGetEmojiQuery(emojiId);

	if (error) {
		return (
			<div className="error accent">
				{error.status}: {error.data.error}
			</div>
		);
	} else if (isLoading) {
		return (
			<div>
				<Loading/>
			</div>
		);
	} else {
		return <EmojiDetail emoji={emoji}/>;
	}
}

function EmojiDetail({emoji}) {
	const [modifyEmoji, modifyResult] = query.useEditEmojiMutation();

	const [isNewCategory, setIsNewCategory] = React.useState(false);

	const [categoryState, _resetCategory, { category }] = useComboBoxInput("category", {defaultValue: emoji.category});

	const [onFileChange, _resetFile, { image, imageURL, imageInfo }] = useFileInput("image", {
		withPreview: true,
		maxSize: 50 * 1024
	});

	function modifyCategory() {
		modifyEmoji({id: emoji.id, category: category.trim()});
	}

	function modifyImage() {
		modifyEmoji({id: emoji.id, image: image});
	}

	React.useEffect(() => {
		if (category != emoji.category && !categoryState.open && !isNewCategory && category.trim().length > 0) {
			console.log("updating to", category);
			modifyEmoji({id: emoji.id, category: category.trim()});
		}
	}, [isNewCategory, category, categoryState.open, emoji.category, emoji.id, modifyEmoji]);

	return (
		<>
			<div className="emoji-header">
				<img src={emoji.url} alt={emoji.shortcode} title={emoji.shortcode}/>
				<div>
					<h2>{emoji.shortcode}</h2>
					<DeleteButton id={emoji.id}/>
				</div>
			</div>

			<div className="left-border">
				<h2>Modify this emoji {modifyResult.isLoading && "(processing..)"}</h2>

				{modifyResult.error && <div className="error">
					{modifyResult.error.status}: {modifyResult.error.data.error}
				</div>}

				<div className="update-category">
					<CategorySelect
						value={category}
						categoryState={categoryState}
						setIsNew={setIsNewCategory}
					>
						<button style={{visibility: (isNewCategory ? "initial" : "hidden")}} onClick={modifyCategory}>
							Create
						</button>
					</CategorySelect>
				</div>

				<div className="update-image">
					<b>Image</b>
					<div className="form-field file">
						<label className="file-input button" htmlFor="image">
							Browse
						</label>
						{imageInfo}
						<input
							className="hidden"
							type="file"
							id="image"
							name="Image"
							accept="image/png,image/gif"
							onChange={onFileChange}
						/>
					</div>

					<button onClick={modifyImage} disabled={image == undefined}>Replace image</button>

					<FakeToot>
						Look at this new custom emoji <img
							className="emoji"
							src={imageURL ?? emoji.url}
							title={`:${emoji.shortcode}:`}
							alt={emoji.shortcode}
						/> isn&apos;t it cool?
					</FakeToot>
				</div>
			</div>
		</>
	);
}

function DeleteButton({id}) {
	// TODO: confirmation dialog?
	const [deleteEmoji, deleteResult] = query.useDeleteEmojiMutation();

	let text = "Delete";
	if (deleteResult.isLoading) {
		text = "Deleting...";
	}

	if (deleteResult.isSuccess) {
		return <Redirect to={base}/>;
	}

	return (
		<button className="danger" onClick={() => deleteEmoji(id)} disabled={deleteResult.isLoading}>{text}</button>
	);
}