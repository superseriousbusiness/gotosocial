/*
	GoToSocial
	Copyright (C) GoToSocial Authors admin@gotosocial.org
	SPDX-License-Identifier: AGPL-3.0-or-later

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

import React from "react";

import { useState } from "react";
import prettierBytes from "prettier-bytes";
import { HookName, HookOpts } from "./types";

export default function useFileInput(
	{ name }: HookName,
	{
		withPreview,
		maxSize,
		initialInfo = "no file selected"
	}: HookOpts
) {
	const [file, setFile] = useState<File>();
	const [imageURL, setImageURL] = useState<string>();
	const [info, setInfo] = useState<React.JSX.Element>();

	function onChange(e: React.ChangeEvent<HTMLInputElement>) {
		const files = e.target.files;
		if (!files) {
			setInfo(undefined);
			return;
		}

		let file = files[0];
		setFile(file);

		if (imageURL) {
			URL.revokeObjectURL(imageURL);
		}
		
		if (withPreview) {
			setImageURL(URL.createObjectURL(file));
		}

		let size = prettierBytes(file.size);
		if (maxSize && file.size > maxSize) {
			size = <span className="error-text">{size}</span>;
		}

		setInfo(
			<>
				{file.name} ({size})
			</>
		);
	}

	function reset() {
		if (imageURL) {
			URL.revokeObjectURL(imageURL);
		}
		setImageURL(undefined);
		setFile(undefined);
		setInfo(undefined);
	}

	const infoComponent = (
		<span className="form-info">
			{info
				? info
				: initialInfo
			}
		</span>
	);

	// Array / Object hybrid, for easier access in different contexts
	return Object.assign([
		onChange,
		reset,
		{
			[name]: file,
			[`${name}URL`]: imageURL,
			[`${name}Info`]: infoComponent,
		}
	], {
		onChange,
		reset,
		name,
		value: file,
		previewValue: imageURL,
		hasChanged: () => file != undefined,
		infoComponent
	});
}
