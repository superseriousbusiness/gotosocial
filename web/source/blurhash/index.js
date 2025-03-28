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

import { decode } from "blurhash";

Array.from(document.getElementsByClassName('blurhash-container')).forEach(blurhashContainer => {
	const hash = blurhashContainer.dataset.blurhashHash;
	const thumbHeight = blurhashContainer.dataset.blurhashHeight;
	const thumbWidth = blurhashContainer.dataset.blurhashWidth;
	const thumbAspect = blurhashContainer.dataset.blurhashAspect;	
	
	/*
		It's very expensive to draw big canvases
		with blurhashes, so use tiny ones, keeping
		aspect ratio of the original thumbnail.
	*/
	var useWidth = 32;
	var useHeight = 32;
	switch (true) {
		case thumbWidth > thumbHeight:
			useHeight = Math.round(useWidth / thumbAspect);
			break;
		case thumbHeight > thumbWidth:
			useWidth = Math.round(useHeight * thumbAspect);
			break;
	}

	const pixels = decode(
		hash,
		useWidth,
		useHeight,
	);

	// Create canvas of appropriate size.
	const canvas = document.createElement("canvas");
	canvas.width = useWidth;
	canvas.height = useHeight;

	// Draw the image data into the canvas.
	const ctx = canvas.getContext("2d");
	const imageData = ctx.createImageData(
		useWidth,
		useHeight,
	);
	imageData.data.set(pixels);
	ctx.putImageData(imageData, 0, 0);

	// Put the canvas inside the container.
	blurhashContainer.appendChild(canvas);
});
