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

const BackButton = require("../../components/back-button");

const query = require("../../lib/query");

const base = "/settings/admin/custom-emoji";

/* We wrap the component to generate formFields with a setter depending on the domain
	 if formFields() is used inside the same component that is re-rendered with their state,
	 inputs get re-created on every change, causing them to lose focus, and bad performance
*/
module.exports = function EmojiDetailWrapped() {
	let [_match, {emojiId}] = useRoute(`${base}/:emojiId`);
	const {currentData: emoji, isLoading, error} = query.useGetEmojiQuery(emojiId);

	return (<>
		{error && <div className="error accent">{error.status}: {error.data.error}</div>}
		{isLoading
			? "Loading..."
			: <EmojiDetail emoji={emoji}/>
		}
	</>);
};

function EmojiDetail({emoji}) {
	if (emoji == undefined) {
		return (<>
			<Link to={base}>
				<a className="button">go back</a>
			</Link>
		</>);
	}

	return (
		<div>
			<h1><BackButton to={base}/> Custom Emoji: {emoji.shortcode}</h1>
			<DeleteButton id={emoji.id}/>
			<p>
				Editing custom emoji isn&apos;t implemented yet.<br/>
				<a target="_blank" rel="noreferrer" href="https://github.com/superseriousbusiness/gotosocial/issues/797">View implementation progress.</a>
			</p>
			<img src={emoji.url} alt={emoji.shortcode} title={`:${emoji.shortcode}:`}/>
		</div>
	);
}

function DeleteButton({id}) {
	// TODO: confirmation dialog?
	const [deleteEmoji, deleteResult] = query.useDeleteEmojiMutation();

	let text = "Delete this emoji";
	if (deleteResult.isLoading) {
		text = "processing...";
	}

	if (deleteResult.isSuccess) {
		return <Redirect to={base}/>;
	}

	return (
		<button onClick={() => deleteEmoji(id)} disabled={deleteResult.isLoading}>{text}</button>
	);
}