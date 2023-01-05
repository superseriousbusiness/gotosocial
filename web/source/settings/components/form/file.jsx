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
const prettierBytes = require("prettier-bytes");

module.exports = function useFileInput({name, _Name}, {
	withPreview,
	maxSize,
	initialInfo = "no file selected"
}) {
	const [file, setFile] = React.useState();
	const [imageURL, setImageURL] = React.useState();
	const [info, setInfo] = React.useState();

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
			setInfo();
		}
	}

	function reset() {
		URL.revokeObjectURL(imageURL);
		setImageURL();
		setFile();
		setInfo();
	}

	return [
		onChange,
		reset,
		{
			[name]: file,
			[`${name}URL`]: imageURL,
			[`${name}Info`]: <span className="form-info">
				{info
					? info
					: initialInfo
				}
			</span>
		}
	];
};