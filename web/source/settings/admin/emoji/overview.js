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
const prettierBytes = require("prettier-bytes");

const FakeToot = require("../../components/fake-toot");
const MutateButton = require("../../components/mutation-button");

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
					<NewEmoji emoji={emoji}/>
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

function useFileInput({withPreview, maxSize}) {
	const [file, setFile] = React.useState();
	const [imageURL, setImageURL] = React.useState();
	const [info, setInfo] = React.useState("no file selected");

	function onChange(e) {
		let file = e.target.files[0];
		setFile(file);

		URL.revokeObjectURL(imageURL);

		if (file != undefined) {
			if (withPreview) {
				setImageURL(URL.createObjectURL(file));
			}
	
			let size = prettierBytes(file.size);
			if (maxSize && file.size > maxSize) {
				size = <span className="error-text">{size}</span>;
			}

			setInfo(<>
				{file.name} ({size})
			</>);
		} else {
			setInfo("no file selected");
		}
	}

	function reset() {
		setFile();
		URL.revokeObjectURL(imageURL);
		setInfo("no file selected");
	}

	return [
		onChange,
		reset,
		{
			file,
			imageURL,
			info: <span>{info}</span>,
		}
	];
}

// TODO: change form field code, maybe look into redux-final-form or similar
// or evaluate if we even need to put most of this in the store
function NewEmoji({emoji}) {
	const emojiCodes = React.useMemo(() => {
		return new Set(emoji.map((e) => e.shortcode));
	}, [emoji]);
	const [addEmoji, result] = query.useAddEmojiMutation();
	const [onFileChange, resetFile, {file, imageURL, info}] = useFileInput({
		withPreview: true,
		maxSize: 50 * 1000
	});

	const [shortcode, setShortcode] = React.useState("");
	const shortcodeRef = React.useRef(null);

	function onShortChange(e) {
		let input = e.target.value;
		setShortcode(input);
		validateShortcode(input);
	}

	function validateShortcode(code) {
		console.log("code: (%s)", code);
		if (emojiCodes.has(code)) {
			shortcodeRef.current.setCustomValidity("Shortcode already in use");
		} else {
			shortcodeRef.current.setCustomValidity("");
		}
		shortcodeRef.current.reportValidity();
	}

	React.useEffect(() => {
		if (shortcode.length == 0) {
			if (file != undefined) {
				let [name, _ext] = file.name.split(".");
				setShortcode(name);
				validateShortcode(name);
			}
		}
	}, [file, shortcode]);

	function uploadEmoji(e) {
		if (e) {
			e.preventDefault();
		}

		Promise.try(() => {
			return addEmoji({
				image: file,
				shortcode
			});
		}).then(() => {
			resetFile();
			setShortcode("");
		});
	}

	let emojiOrShortcode = `:${shortcode}:`;

	if (imageURL != undefined) {
		emojiOrShortcode = <img
			className="emoji"
			src={imageURL}
			title={`:${shortcode}:`}
			alt={shortcode}
		/>;
	}

	return (
		<div>
			<h2>Add new custom emoji</h2>

			<FakeToot>
				Look at this new custom emoji {emojiOrShortcode} isn&apos;t it cool?
			</FakeToot>

			<form onSubmit={uploadEmoji} className="form-flex">
				<div className="form-field file">
					<label className="file-input button" htmlFor="image">
						Browse
					</label>
					{info}
					<input
						className="hidden"
						type="file"
						id="image"
						name="Image"
						accept="image/png,image/gif"
						onChange={onFileChange}
					/>
				</div>

				<div className="form-field text">
					<label htmlFor="shortcode">
						Shortcode, must be unique among the instance's local emoji
					</label>
					<input
						type="text"
						id="shortcode"
						name="Shortcode"
						ref={shortcodeRef}
						onChange={onShortChange}
						value={shortcode}
					/>
				</div>
				
				<MutateButton text="Upload emoji" result={result}/>
			</form>
		</div>
	);
}