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
const {Switch, Route, Link, Redirect, useRoute, useLocation} = require("wouter");

const Submit = require("../components/submit");
const FakeToot = require("../components/fake-toot");
const { formFields } = require("../components/form-fields");

const api = require("../lib/api");
const adminActions = require("../redux/reducers/admin").actions;
const submit = require("../lib/submit");

const base = "/settings/admin/custom-emoji";

module.exports = function CustomEmoji() {
	return (
		<Switch>
			<Route path={`${base}/:emojiId`}>
				<EmojiDetailWrapped />
			</Route>
			<EmojiOverview />
		</Switch>
	);
};

function EmojiOverview() {
	const dispatch = Redux.useDispatch();
	const [loaded, setLoaded] = React.useState(false);

	const [errorMsg, setError] = React.useState("");

	React.useEffect(() => {
		if (!loaded) {
			Promise.try(() => {
				return dispatch(api.admin.fetchCustomEmoji());
			}).then(() => {
				setLoaded(true);
			}).catch((e) => {
				setLoaded(true);
				setError(e.message);
			});
		}
	}, []);

	if (!loaded) {
		return (
			<>
				<h1>Custom Emoji</h1>
				Loading...
			</>
		);
	}

	return (
		<>
			<h1>Custom Emoji</h1>
			<EmojiList/>
			<NewEmoji/>
			{errorMsg.length > 0 && 
				<div className="error accent">{errorMsg}</div>
			}
		</>
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

			<FakeToot content="bazinga">
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

function EmojiList() {
	const emoji = Redux.useSelector((state) => state.admin.emoji);

	return (
		<div>
			<h2>Overview</h2>
			<div className="list emoji-list">
				{Object.entries(emoji).map(([category, entries]) => {
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
						// <Link key={e.static_url} to={`${base}/${e.shortcode}`}>
						<Link key={e.static_url} to={`${base}`}>
							<a>
								<img src={e.static_url} alt={e.shortcode} title={`:${e.shortcode}:`}/>
							</a>
						</Link>
					);
				})}
			</div>
		</div>
	);
}

function EmojiDetailWrapped() {
	/* We wrap the component to generate formFields with a setter depending on the domain
		 if formFields() is used inside the same component that is re-rendered with their state,
		 inputs get re-created on every change, causing them to lose focus, and bad performance
	*/
	let [_match, {emojiId}] = useRoute(`${base}/:emojiId`);

	function alterEmoji([key, val]) {
		return adminActions.updateDomainBlockVal([emojiId, key, val]);
	}

	const fields = formFields(alterEmoji, (state) => state.admin.blockedInstances[emojiId]);

	return <EmojiDetail id={emojiId} Form={fields} />;
}

function EmojiDetail({id, Form}) {
	return (
		"Not implemented yet"
	);
}