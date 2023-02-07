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
	useTextInput,
	useComboBoxInput,
	useCheckListInput
} = require("../../../lib/form");

const useFormSubmit = require("../../../lib/form/submit");

const CheckList = require("../../../components/check-list");
const { CategorySelect } = require('../category-select');

const { TextInput } = require("../../../components/form/inputs");
const MutationButton = require("../../../components/form/mutation-button");
const { Error } = require("../../../components/error");

module.exports = function ParseFromToot({ emojiCodes }) {
	const [searchStatus, result] = query.useSearchStatusForEmojiMutation();

	const [onURLChange, _resetURL, { url }] = useTextInput("url");

	function submitSearch(e) {
		e.preventDefault();
		if (url.trim().length != 0) {
			searchStatus(url);
		}
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
						<button disabled={result.isLoading}>
							<i className={[
								"fa fa-fw",
								(result.isLoading
									? "fa-refresh fa-spin"
									: "fa-search")
							].join(" ")} aria-hidden="true" title="Search" />
							<span className="sr-only">Search</span>
						</button>
					</div>
				</div>
			</form>
			<SearchResult result={result} localEmojiCodes={emojiCodes} />
		</div>
	);
};

function SearchResult({ result, localEmojiCodes }) {
	const { error, data, isSuccess, isError } = result;

	if (!(isSuccess || isError)) {
		return null;
	}

	if (error == "NONE_FOUND") {
		return "No results found";
	} else if (error == "LOCAL_INSTANCE") {
		return <b>This is a local user/toot, all referenced emoji are already on your instance</b>;
	} else if (error != undefined) {
		return <Error error={result.error} />;
	}

	if (data.list.length == 0) {
		return <b>This {data.type == "statuses" ? "toot" : "account"} doesn't use any custom emoji</b>;
	}

	return (
		<CopyEmojiForm
			localEmojiCodes={localEmojiCodes}
			type={data.type}
			domain={data.domain}
			emojiList={data.list}
		/>
	);
}

function CopyEmojiForm({ localEmojiCodes, type, emojiList }) {
	const form = {
		selectedEmoji: useCheckListInput("selectedEmoji", {
			entries: emojiList,
			uniqueKey: "id"
		}),
		category: useComboBoxInput("category")
	};

	const [formSubmit, result] = useFormSubmit(
		form,
		query.usePatchRemoteEmojisMutation(),
		{
			changedOnly: false,
			onFinish: ({ data }) => {
				if (data != undefined) {
					form.selectedEmoji.updateMultiple(
						// uncheck all successfully processed emoji
						data.map(([id]) => [id, { checked: false }])
					);
				}
			}
		}
	);

	const buttonsInactive = form.selectedEmoji.someSelected
		? {}
		: {
			disabled: true,
			title: "No emoji selected, cannot perform any actions"
		};

	const checkListExtraProps = React.useCallback(() => ({ localEmojiCodes }), [localEmojiCodes]);

	return (
		<div className="parsed">
			<span>This {type == "statuses" ? "toot" : "account"} uses the following custom emoji, select the ones you want to copy/disable:</span>
			<form onSubmit={formSubmit}>
				<CheckList
					field={form.selectedEmoji}
					EntryComponent={EmojiEntry}
					getExtraProps={checkListExtraProps}
				/>

				<CategorySelect
					field={form.category}
				/>

				<div className="action-buttons row">
					<MutationButton name="copy" label="Copy to local emoji" result={result} showError={false} {...buttonsInactive} />
					<MutationButton name="disable" label="Disable" result={result} className="button danger" showError={false} {...buttonsInactive} />
				</div>
				{result.error && (
					Array.isArray(result.error)
						? <ErrorList errors={result.error} />
						: <Error error={result.error} />
				)}
			</form>
		</div>
	);
}

function ErrorList({ errors }) {
	return (
		<div className="error">
			One or multiple emoji failed to process:
			{errors.map(([shortcode, err]) => (
				<div key={shortcode}>
					<b>{shortcode}:</b> {err}
				</div>
			))}
		</div>
	);
}

function EmojiEntry({ entry: emoji, onChange, extraProps: { localEmojiCodes } }) {
	const shortcodeField = useTextInput("shortcode", {
		defaultValue: emoji.shortcode,
		validator: function validateShortcode(code) {
			return (emoji.checked && localEmojiCodes.has(code))
				? "Shortcode already in use"
				: "";
		}
	});

	React.useEffect(() => {
		if (emoji.valid != shortcodeField.valid) {
			onChange({ valid: shortcodeField.valid });
		}
	}, [onChange, emoji.valid, shortcodeField.valid]);

	React.useEffect(() => {
		shortcodeField.validate();
		// only need this update if it's the emoji.checked that updated, not shortcodeField
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [emoji.checked]);

	return (
		<>
			<img className="emoji" src={emoji.url} title={emoji.shortcode} />

			<TextInput
				field={shortcodeField}
				onChange={(e) => {
					shortcodeField.onChange(e);
					onChange({ shortcode: e.target.value, checked: true });
				}}
			/>
		</>
	);
}