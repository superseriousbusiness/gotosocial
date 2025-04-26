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

/*
	WHAT SHOULD GO IN THIS FILE?

	This script is loaded just before the end of the HTML body, so
	put stuff in here that should be run *before* the user sees the page.
	So, stuff that shifts the layout or causes elements to jump around.
*/

import { decode } from "blurhash";

const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)');

// Adjust object-position of any image that has a focal point set.
document.querySelectorAll("img[data-object-position]").forEach(img => {
	img.style["object-position"] = img.dataset.objectPosition;
});

// Generate a blurhash canvas for each image for
// each blurhash container and put it in the summary.
Array.from(document.getElementsByClassName('blurhash-container')).forEach(blurhashContainer => {
	const hash = blurhashContainer.dataset.blurhashHash;
	const thumbHeight = blurhashContainer.dataset.blurhashHeight;
	const thumbWidth = blurhashContainer.dataset.blurhashWidth;
	const thumbAspect = blurhashContainer.dataset.blurhashAspect;	
	const objectPosition = blurhashContainer.dataset.blurhashObjectPosition;
	
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

	// Set object-position css property on
	// the canvas if it's set on the container.
	if (objectPosition) {
		canvas.style["object-position"] = objectPosition;
	}

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

// Change the spoiler / content warning boxes from generic
// "toggle visibility" to show/hide depending on state,
// and add keyboard functionality to spoiler buttons.
function dynamicSpoiler(className, updateFunc) {
	Array.from(document.getElementsByClassName(className)).forEach((spoiler) => {
		const update = updateFunc(spoiler);
		if (update) {
			update();
			spoiler.addEventListener("toggle", update);
		}
	});
}
dynamicSpoiler("text-spoiler", (details) => {
	const summary = details.children[0];
	const button = details.querySelector(".button");

	// Use button *instead of summary*
	// to toggle post visibility.
	summary.tabIndex = "-1";
	button.tabIndex = "0";
	button.setAttribute("aria-role", "button");
	button.onclick = (e) => {
		e.preventDefault();
		return details.hasAttribute("open")
			? details.removeAttribute("open")
			: details.setAttribute("open", "");
	};

	// Let enter also trigger the button
	// (for those using keyboard to navigate).
	button.addEventListener("keydown", (e) => {
		if (e.key === "Enter") {
			e.preventDefault();
			button.click();
		}
	});

	// Change button text depending on
	// whether spoiler is open or closed rn.
	return () => {
		button.textContent = details.open
			? "Show less"
			: "Show more";
	};
});
dynamicSpoiler("media-spoiler", (details) => {
	const summary = details.children[0];
	const button = details.querySelector(".eye.button");
	const video = details.querySelector(".plyr-video");
	const loopingAuto = !reduceMotion.matches && video != null && video.classList.contains("gifv");

	// Use button *instead of summary*
	// to toggle media visibility.
	summary.tabIndex = "-1";
	button.tabIndex = "0";
	button.setAttribute("aria-role", "button");
	button.onclick = (e) => {
		e.preventDefault();
		return details.hasAttribute("open")
			? details.removeAttribute("open")
			: details.setAttribute("open", "");
	};

	// Let enter also trigger the button
	// (for those using keyboard to navigate).
	button.addEventListener("keydown", (e) => {
		if (e.key === "Enter") {
			e.preventDefault();
			button.click();
		}
	});

	return () => {
		if (details.open) {
			button.setAttribute("aria-label", "Hide media");
		} else {
			button.setAttribute("aria-label", "Show media");
			if (video && !loopingAuto) {
				video.pause();
			}
		}
	};
});

// Reformat time text to browser locale.
// Define + reuse one DateTimeFormat (cheaper).
const dateTimeFormat = Intl.DateTimeFormat(
	undefined,
	{
		year: 'numeric',
		month: 'short',
		day: '2-digit',
		hour: '2-digit',
		minute: '2-digit',
		hour12: false
	},
);
Array.from(document.getElementsByTagName('time')).forEach(timeTag => {
	const datetime = timeTag.getAttribute('datetime');
	const currentText = timeTag.textContent.trim();
	// Only format if current text contains precise time.
	if (currentText.match(/\d{2}:\d{2}/)) {
		const date = new Date(datetime);
		timeTag.textContent = dateTimeFormat.format(date);
	}
});
