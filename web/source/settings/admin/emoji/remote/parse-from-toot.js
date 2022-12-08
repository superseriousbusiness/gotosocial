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
const syncpipe = require("syncpipe");

const {
	useTextInput,
	useComboBoxInput
} = require("../../../components/form");

const { CategorySelect } = require('../category-select');

const query = require("../../../lib/query");

module.exports = function ParseFromToot({emojiCodes}) {
	const [searchStatus, { data }] = query.useSearchStatusForEmojiMutation();

	const [onURLChange, _resetURL, { url }] = useTextInput("url", {defaultValue: "http://localhost:8080/@admin/statuses/01GKPHE83MV6D2SF3ZFKEJ8Y4F"});

	const status = data?.statuses?.[0];
	const emojiList = status?.emojis;

	function submitSearch(e) {
		e.preventDefault();
		searchStatus(url);
	}

	return (
		<div className="parse-emoji">
			<h2>Steal this look</h2>
			<form onSubmit={submitSearch}>
				<div className="form-field text">
					<label htmlFor="shortcode">
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
						<button><i className="fa fa-search" aria-hidden="true"></i></button>
					</div>
				</div>
			</form>
			{status && 
				/* FIXME: remove || true, remove ?? "local" */
				(status.account.acct.includes("@") || true
					? <CopyEmojiForm localEmojiCodes={emojiCodes} emojiList={emojiList} domain={status.account.acct.split("@")[1] ?? "local"}/>
					: <b>This is a local toot, all emoji are already on your instance</b>
				)
			}
		</div>
	);
};

function makeEmojiState(emojiList, checked) {
	/* Return a new object, with a key for every emoji's shortcode,
	   And a value for it's checkbox `checked` state.
	 */
	return syncpipe(emojiList, [
		(_) => _.map((emoji) => [emoji.shortcode, {
			checked,
			valid: true
		}]),
		(_) => Object.fromEntries(_)
	]);
}

function updateEmojiState(emojiState, checked) {
	/* Create a new object with all emoji entries' checked state updated */
	return syncpipe(emojiState, [
		(_) => Object.entries(emojiState),
		(_) => _.map(([key, val]) => [key, {
			...val,
			checked
		}]),
		(_) => Object.fromEntries(_)
	]);
}

function CopyEmojiForm({localEmojiCodes, emojiList, domain}) {
	const [patchRemoteEmojis, _patchResult] = query.usePatchRemoteEmojisMutation();
	const [err, setError] = React.useState();

	const toggleAllRef = React.useRef(null);
	const [toggleAllState, setToggleAllState] = React.useState(0);
	const [emojiState, setEmojiState] = React.useState(makeEmojiState(emojiList, false));
	const [someSelected, setSomeSelected] = React.useState(false);

	const [categoryState, resetCategory, { category }] = useComboBoxInput("category");

	React.useEffect(() => {
		if (emojiList != undefined) {
			setEmojiState(makeEmojiState(emojiList, false));
		}
	}, [emojiList]);

	React.useEffect(() => {
		/* Updates (un)check all checkbox, based on shortcode checkboxes
		   Can be 0 (not checked), 1 (checked) or 2 (indeterminate)
		 */
		if (toggleAllRef.current == null) {
			return;
		}

		let values = Object.values(emojiState);
		/* one or more boxes are checked */
		let some = values.some((v) => v.checked);

		let all = false;
		if (some) {
			/* there's not at least one unchecked box */
			all = !values.some((v) => v.checked == false);
		}

		setSomeSelected(some);

		if (some && !all) {
			setToggleAllState(2);
			toggleAllRef.current.indeterminate = true;
		} else {
			setToggleAllState(all ? 1 : 0);
			toggleAllRef.current.indeterminate = false;
		}
	}, [emojiState, toggleAllRef]);

	function updateEmoji(shortcode, value) {
		setEmojiState({
			...emojiState,
			[shortcode]: {
				...emojiState[shortcode],
				...value
			}
		});
	}

	function toggleAll(e) {
		let selectAll = e.target.checked;

		if (toggleAllState == 2) { // indeterminate
			selectAll = false;
		}

		setEmojiState(updateEmojiState(emojiState, selectAll));
		setToggleAllState(selectAll);
	}

	function submit(action) {
		Promise.try(() => {
			setError(null);
			const selectedShortcodes = syncpipe(emojiState, [
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
				list: selectedShortcodes
			}).unwrap();
		}).then(() => {
			setEmojiState(makeEmojiState(emojiList, false));
			resetCategory();
		}).catch((e) => {
			if (Array.isArray(e)) {
				setError(e.map(([shortcode, msg]) => (
					<div key={shortcode}>
						{shortcode}: <span style={{fontWeight: "initial"}}>{msg}</span>
					</div>
				)));
			} else {
				setError(e);
			}
		});
	}

	return (
		<div className="parsed">
			<span>This toot includes the following custom emoji, select the ones you want to act on:</span>
			<div className="emoji-list">
				<label className="header">
					<input
						ref={toggleAllRef}
						type="checkbox"
						onChange={toggleAll}
						checked={toggleAllState === 1}
					/> All
				</label>
				{emojiList.map((emoji) => (
					<EmojiEntry
						key={emoji.shortcode}
						emoji={emoji}
						localEmojiCodes={localEmojiCodes}
						updateEmoji={(value) => updateEmoji(emoji.shortcode, value)}
						checked={emojiState[emoji.shortcode].checked}
					/>
				))}
			</div>

			<CategorySelect
				value={category}
				categoryState={categoryState}
			/>

			<div className="action-buttons row">
				<button disabled={!someSelected} onClick={() => submit("copy")}>Copy to local emoji</button>
				<button disabled={!someSelected} onClick={() => submit("disable")} className="danger">Disable</button>
			</div>
			{err && <div className="error">
				{err}
			</div>}
		</div>
	);
}

function EmojiEntry({emoji, localEmojiCodes, updateEmoji, checked}) {
	const [onShortcodeChange, _resetShortcode, { shortcode, shortcodeRef, shortcodeValid}] = useTextInput("shortcode", {
		defaultValue: emoji.shortcode,
		validator: function validateShortcode(code) {
			return (checked && localEmojiCodes.has(code))
				? "Shortcode already in use"
				: "";
		}
	});

	React.useEffect(() => {
		updateEmoji({valid: shortcodeValid});
		/* eslint-disable-next-line react-hooks/exhaustive-deps */
	}, [shortcodeValid]);

	return (
		<label key={emoji.shortcode} className="row">
			<input
				type="checkbox"
				onChange={(e) => updateEmoji({checked: e.target.checked})}
				checked={checked}
			/>
			<img className="emoji" src={emoji.url} title={emoji.shortcode} />

			<input
				type="text"
				id="shortcode"
				name="Shortcode"
				ref={shortcodeRef}
				onChange={(e) => {
					onShortcodeChange(e);
					updateEmoji({shortcode: e.target.value, checked: true});
				}}
				value={shortcode}
			/>
		</label>
	);
}