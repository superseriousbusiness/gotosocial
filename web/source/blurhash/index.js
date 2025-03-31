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

// Generate a blurhash canvas for each image for
// each blurhash container and put it in the summary.
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

// Add a smooth transition from blurhash
// to image for each sensitive image.
Array.from(document.getElementsByTagName('img')).forEach(img => {
	if (!img.dataset.blurhashHash) {
		// Has no blurhash,
		// can't transition smoothly.
		return;
	}

	if (img.dataset.sensitive !== "true") {
		// Not sensitive, smooth
		// transition doesn't matter.
		return;
	}

	if (img.complete) {
		// Image already loaded,
		// don't stub it with anything.
		return;
	}
	
	const parentSlide = img.closest(".photoswipe-slide");
	if (!parentSlide) {
		// Parent slide was nil,
		// can't do anything.
		return;
	}

	const blurhashContainer = document.querySelector("div[data-blurhash-hash=\"" + img.dataset.blurhashHash + "\"]");
	if (!blurhashContainer) {
		// Blurhash div was nil,
		// can't do anything.
		return;
	}

	const canvas = blurhashContainer.children[0];
	if (!canvas) {
		// Canvas was nil,
		// can't do anything.
		return;
	} 

	// "Replace" the hidden img with a canvas
	// that will show initially when it's clicked.
	const clone = canvas.cloneNode(true);
	clone.getContext("2d").drawImage(canvas, 0, 0);
	parentSlide.prepend(clone);
	img.className = img.className + " hidden";

	// Add a listener so that when the spoiler
	// is opened, loading of the image begins.
	const parentSummary = img.closest(".media-spoiler");
	parentSummary.addEventListener("toggle", (_) => {
		if (parentSummary.hasAttribute("open") && !img.complete) {
			img.loading = "eager";
		}
	});

	// Add a callback that triggers
	// when image loading is complete.
	img.addEventListener("load", () => {
		// Show the image now that it's loaded.
		img.className = img.className.replace(" hidden", "");
		
		// Reset the lazy loading tag to its initial
		// value. This doesn't matter too much since
		// it's already loaded but it feels neater.
		img.loading = "lazy";

		// Remove the temporary blurhash
		// canvas so only the image shows.
		const canvas = parentSlide.getElementsByTagName("canvas")[0];
		if (canvas) {
			canvas.remove();
		}
	});
});
