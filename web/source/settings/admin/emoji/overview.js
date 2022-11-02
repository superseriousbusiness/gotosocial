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

const Promise = require("bluebird");
const React = require("react");
const Redux = require("react-redux");
const {Link} = require("wouter");
const defaultValue = require('default-value');

const Submit = require("../../components/submit");
const FakeToot = require("../../components/fake-toot");
const { formFields } = require("../../components/form-fields");

const api = require("../../lib/api");
const adminActions = require("../../redux/reducers/admin").actions;
const submit = require("../../lib/submit");

const query = require("../../lib/query");

const base = "/settings/admin/custom-emoji";

module.exports = function EmojiOverview() {
	const {
		data: emoji,
		isLoading,
		error
	} = query.useGetAllEmojiQuery({filter: "domain:local"});

	return (
		<>
			<h1>Custom Emoji</h1>
			{error && 
				<div className="error accent">{error}</div>
			}
			{isLoading
				? "Loading..."
				: <>
					<EmojiList emoji={emoji}/>
					<NewEmoji/>
				</>
			}
		</>
	);
};

function EmojiList({emoji}) {
	const byCategory = React.useMemo(() => {
		const categories = {};

		emoji.forEach((emoji) => {
			let cat = defaultValue(emoji.category, "Unsorted");
			categories[cat] = defaultValue(categories[cat], []);
			categories[cat].push(emoji);
		});

		return categories;
	}, [emoji]);
	
	return (
		<div>
			<h2>Overview</h2>
			<div className="list emoji-list">
				{emoji.length == 0 && "No local emoji yet"}
				{Object.entries(byCategory).map(([category, entries]) => {
					return <EmojiCategory key={category} category={category} entries={entries}/>;
				})}
			</div>
		</div>
	);
}

function EmojiCategory({category, entries}) {
	return (
		<div className="entry">
			<b>{category}</b>
			<div className="emoji-group">
				{entries.map((e) => {
					return (
						<Link key={e.id} to={`${base}/${e.id}`}>
							{/* <Link key={e.static_url} to={`${base}`}> */}
							<a>
								<img src={e.url} alt={e.shortcode} title={`:${e.shortcode}:`}/>
							</a>
						</Link>
					);
				})}
			</div>
		</div>
	);
}

const NewEmojiForm = formFields(adminActions.updateNewEmojiVal, (state) => state.admin.newEmoji);
function NewEmoji() {
	const dispatch = Redux.useDispatch();
	const newEmojiForm = Redux.useSelector((state) => state.admin.newEmoji);

	const [errorMsg, setError] = React.useState("");
	const [statusMsg, setStatus] = React.useState("");

	const uploadEmoji = submit(
		() => dispatch(api.admin.newEmoji()),
		{
			setStatus, setError,
			onSuccess: function() {
				URL.revokeObjectURL(newEmojiForm.image);
				return Promise.all([
					dispatch(adminActions.updateNewEmojiVal(["image", undefined])),
					dispatch(adminActions.updateNewEmojiVal(["imageFile", undefined])),
					dispatch(adminActions.updateNewEmojiVal(["shortcode", ""])),
				]);
			}
		}
	);

	React.useEffect(() => {
		if (newEmojiForm.shortcode.length == 0) {
			if (newEmojiForm.imageFile != undefined) {
				let [name, ext] = newEmojiForm.imageFile.name.split(".");
				dispatch(adminActions.updateNewEmojiVal(["shortcode", name]));
			}
		}
	});

	let emojiOrShortcode = `:${newEmojiForm.shortcode}:`;

	if (newEmojiForm.image != undefined) {
		emojiOrShortcode = <img
			className="emoji"
			src={newEmojiForm.image}
			title={`:${newEmojiForm.shortcode}:`}
			alt={newEmojiForm.shortcode}
		/>;
	}

	return (
		<div>
			<h2>Add new custom emoji</h2>

			<FakeToot>
				Look at this new custom emoji {emojiOrShortcode} isn&apos;t it cool?
			</FakeToot>

			<NewEmojiForm.File
				id="image"
				name="Image"
				fileType="image/png,image/gif"
				showSize={true}
				maxSize={50 * 1000}
			/>

			<NewEmojiForm.TextInput
				id="shortcode"
				name="Shortcode (without : :), must be unique on the instance"
				placeHolder="blobcat"
			/>

			<Submit onClick={uploadEmoji} label="Upload" errorMsg={errorMsg} statusMsg={statusMsg} />
		</div>
	);
}