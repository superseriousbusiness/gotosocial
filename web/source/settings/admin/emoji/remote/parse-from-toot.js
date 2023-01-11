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

const Promise = require("bluebird");
const React = require("react");
const Redux = require("react-redux");
const syncpipe = require("syncpipe");

const {
	useTextInput,
	useComboBoxInput,
	useCheckListInput
} = require("../../../lib/form");

const CheckList = require("../../../components/check-list");
const { CategorySelect } = require('../category-select');

const query = require("../../../lib/query");
const Loading = require("../../../components/loading");

module.exports = function ParseFromToot({ emojiCodes }) {
	const [searchStatus, { data, isLoading, isSuccess, error }] = query.useSearchStatusForEmojiMutation();
	const instanceDomain = Redux.useSelector((state) => (new URL(state.oauth.instance).host));

	const [onURLChange, _resetURL, { url }] = useTextInput("url");

	const searchResult = React.useMemo(() => {
		if (!isSuccess) {
			return null;
		}

		if (data.type == "none") {
			return "No results found";
		}

		if (data.domain == instanceDomain) {
			return <b>This is a local user/toot, all referenced emoji are already on your instance</b>;
		}

		if (data.list.length == 0) {
			return <b>This {data.type == "statuses" ? "toot" : "account"} doesn't use any custom emoji</b>;
		}

		return (
			<CopyEmojiForm
				localEmojiCodes={emojiCodes}
				type={data.type}
				domain={data.domain}
				emojiList={data.list}
			/>
		);
	}, [isSuccess, data, instanceDomain, emojiCodes]);

	function submitSearch(e) {
		e.preventDefault();
		searchStatus(url);
	}

	return (
		<div className="parse-emoji">
			<h2>Steal this look</h2>
			<form onSubmit={submitSearch}>
				<div className="form-field text">
					<label htmlFor="url">
						Link to a toot:
					</label>
					<div className="row">
						<input
							type="text"
							id="url"
							name="url"
							onChange={onURLChange}
							value={url}
						/>
						<button className="button-inline" disabled={isLoading}>
							<i className={[
								"fa",
								(isLoading
									? "fa-refresh fa-spin"
									: "fa-search")
							].join(" ")} aria-hidden="true" title="Search" />
							<span className="sr-only">Search</span>
						</button>
					</div>
					{isLoading && <Loading />}
					{error && <div className="error">{error.data.error}</div>}
				</div>
			</form>
			{searchResult}
		</div>
	);
};

function CopyEmojiForm({ localEmojiCodes, type, domain, emojiList }) {
	const [patchRemoteEmojis, patchResult] = query.usePatchRemoteEmojisMutation();
	const [err, setError] = React.useState();

	const emojiCheckList = useCheckListInput("selectedEmoji", {
		entries: emojiList,
		uniqueKey: "shortcode"
	});

	const [categoryState, resetCategory, { category }] = useComboBoxInput("category");

	function submit(action) {
		Promise.try(() => {
			setError(null);
			const selectedShortcodes = syncpipe(emojiCheckList.value, [
				(_) => Object.entries(_),
				(_) => _.filter(([_shortcode, entry]) => entry.checked),
				(_) => _.map(([shortcode, entry]) => {
					if (action == "copy" && !entry.valid) {
						throw `One or more selected emoji have non-unique shortcodes (${shortcode}), unselect them or pick a different local shortcode`;
					}
					return {
						shortcode,
						localShortcode: entry.shortcode
					};
				})
			]);

			return patchRemoteEmojis({
				action,
				domain,
				list: selectedShortcodes,
				category
			}).unwrap();
		}).then(() => {
			emojiCheckList.reset();
			resetCategory();
		}).catch((e) => {
			if (Array.isArray(e)) {
				setError(e.map(([shortcode, msg]) => (
					<div key={shortcode}>
						{shortcode}: <span style={{ fontWeight: "initial" }}>{msg}</span>
					</div>
				)));
			} else {
				setError(e);
			}
		});
	}

	return (
		<div className="parsed">
			<span>This {type == "statuses" ? "toot" : "account"} uses the following custom emoji, select the ones you want to copy/disable:</span>
			<CheckList
				field={emojiCheckList}
				Component={EmojiEntry}
				localEmojiCodes={localEmojiCodes}
			/>
			{/* {emojiList.map((emoji) => (
				<EmojiEntry
					key={emoji.shortcode}
					emoji={emoji}
					localEmojiCodes={localEmojiCodes}
					updateEmoji={(value) => updateEmoji(emoji.shortcode, value)}
					checked={emojiState[emoji.shortcode].checked}
				/>
			))} */}

			<CategorySelect
				value={category}
				categoryState={categoryState}
			/>

			<div className="action-buttons row">
				<button disabled={!emojiCheckList.someSelected} onClick={() => submit("copy")}>{patchResult.isLoading ? "Processing..." : "Copy to local emoji"}</button>
				<button disabled={!emojiCheckList.someSelected} onClick={() => submit("disable")} className="danger">{patchResult.isLoading ? "Processing..." : "Disable"}</button>
			</div>
			{err && <div className="error">
				{err}
			</div>}
			{patchResult.isSuccess && <div>
				Action applied to {patchResult.data.length} emoji
			</div>}
		</div>
	);
}

function EmojiEntry({ entry: emoji, localEmojiCodes, onChange }) {
	const [onShortcodeChange, _resetShortcode, { shortcode, shortcodeRef, shortcodeValid }] = useTextInput("shortcode", {
		defaultValue: emoji.shortcode,
		validator: function validateShortcode(code) {
			return (emoji.checked && localEmojiCodes.has(code))
				? "Shortcode already in use"
				: "";
		}
	});

	React.useEffect(() => {
		onChange({ valid: shortcodeValid });
		/* eslint-disable-next-line react-hooks/exhaustive-deps */
	}, [shortcodeValid]);

	return (
		<>
			<img className="emoji" src={emoji.url} title={emoji.shortcode} />

			<input
				type="text"
				id="shortcode"
				name="Shortcode"
				ref={shortcodeRef}
				onChange={(e) => {
					onShortcodeChange(e);
					onChange({ shortcode: e.target.value, checked: true });
				}}
				value={shortcode}
			/>
		</>
	);
}