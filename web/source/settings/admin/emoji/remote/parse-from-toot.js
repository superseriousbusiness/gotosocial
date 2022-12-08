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
const syncpipe = require("syncpipe");

const {
	useTextInput
} = require("../../../components/form");

const query = require("../../../lib/query");

function makeEmojiState(emojiList, checked) {
	/* Return a new object, with a key for every emoji's shortcode,
	   And a value for it's checkbox `checked` state.
	 */
	return syncpipe(emojiList, [
		(_) => _.map((emoji) => [emoji.shortcode, checked]),
		(_) => Object.fromEntries(_)
	]);
}

module.exports = function ParseFromToot() {
	const [searchStatus, { data }] = query.useSearchStatusForEmojiMutation();

	const toggleAllRef = React.useRef(null);
	const [toggleAllState, setToggleAllState] = React.useState(0);
	const [emojiChecked, setEmojiChecked] = React.useState({});

	const [onURLChange, _resetURL, { url }] = useTextInput("url");

	const emojiList = data?.statuses?.[0]?.emojis;

	React.useEffect(() => {
		if (emojiList != undefined) {
			setEmojiChecked(makeEmojiState(emojiList, false));
		}
	}, [emojiList]);

	React.useEffect(() => {
		/* Updates (un)check all checkbox, based on shortcode checkboxes
		   Can be 0 (not checked), 1 (checked) or 2 (indeterminate)
		 */
		if (toggleAllRef.current == null) {
			return;
		}

		let values = Object.values(emojiChecked);
		/* one or more boxes are checked */
		let some = values.some((v) => v);
		/* there's not at least one unchecked box */
		let all = !values.some((v) => v == false);

		if (some && !all) {
			setToggleAllState(2);
			toggleAllRef.current.indeterminate = true;
		} else {
			setToggleAllState(all ? 1 : 0);
			toggleAllRef.current.indeterminate = false;
		}
	}, [emojiChecked, toggleAllRef]);

	function submitSearch(e) {
		e.preventDefault();
		searchStatus(url);
	}

	function toggleSpecific(shortcode, checked) {
		setEmojiChecked({
			...emojiChecked,
			[shortcode]: checked
		});
	}

	function toggleAll(e) {
		let selectAll = e.target.checked;

		if (toggleAllState == 2) { // indeterminate
			selectAll = false;
		}

		setEmojiChecked(makeEmojiState(emojiList, selectAll));
		setToggleAllState(selectAll);
	}

	return (
		<div className="parse-emoji">
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
			<div className="parsed">
				{emojiList && <>
					<span>This toot includes the following custom emoji:</span>
					<div className="emoji-list">
						<label className="header">
							<input
								ref={toggleAllRef}
								type="checkbox"
								onChange={toggleAll}
								checked={toggleAllState === 1}
							/> {toggleAllState == 0 ? "select" : "unselect"} all
						</label>
						{emojiList.map((emoji) => (
							<label key={emoji.shortcode} className="row">
								<input
									type="checkbox"
									onChange={(e) => toggleSpecific(emoji.shortcode, e.target.checked)}
									checked={emojiChecked[emoji.shortcode] ?? false}
								/>
								<img className="emoji" src={emoji.url} title={emoji.shortcode} />
								<span>{emoji.shortcode}</span>
							</label>
						))}
					</div>
					<div className="row">
						<button>Copy to local emoji</button>
						<button className="danger">Disable</button>
					</div>
				</>}
			</div>
		</div>
	);
}