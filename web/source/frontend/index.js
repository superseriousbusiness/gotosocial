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

"use strict";

window.HELP_IMPROVE_VIDEOJS = false;

const Photoswipe = require("photoswipe/dist/umd/photoswipe.umd.min.js");
const PhotoswipeLightbox = require("photoswipe/dist/umd/photoswipe-lightbox.umd.min.js");
const PhotoswipeCaptionPlugin = require("photoswipe-dynamic-caption-plugin").default;
const Plyr = require("plyr");

let [_, _user, type, id] = window.location.pathname.split("/");
if (type == "statuses") {
	let firstStatus = document.getElementsByClassName("thread")[0].children[0];
	if (firstStatus.id != id) {
		document.getElementById(id).scrollIntoView();
	}
}

const lightbox = new PhotoswipeLightbox({
	gallery: '.photoswipe-gallery',
	children: '.photoswipe-slide',
	pswpModule: Photoswipe,
});

new PhotoswipeCaptionPlugin(lightbox, {
	type: 'auto',
	captionContent(slide) {
		return slide.data.alt;
	}
});

lightbox.addFilter('itemData', (item) => {
	const el = item.element;
	console.log(el);
	if (el && el.classList.contains("plyr-video")) {
		const parentNode = el._plyrContainer.parentNode;

		return {
			alt: el.getAttribute("alt"),
			_video: {
				open(c) {
					c.appendChild(el._plyrContainer);
				},
				close() {
					parentNode.appendChild(el._plyrContainer);
				}
			},
			width: parseInt(el.dataset.pswpWidth),
			height: parseInt(el.dataset.pswpHeight)
		};
	}
	return item;
});

lightbox.on("contentActivate", (e) => {
	const { content } = e;
	if (content.data._video != undefined) {
		console.log("activate", e);
		content.data._video.open(content.element);
	}
});

lightbox.on("contentDeactivate", (e) => {
	const { content } = e;
	if (content.data._video != undefined) {
		console.log("deactivate", e);
		content.data._video.close();
	}
});

lightbox.on("close", function () {
	console.log("close", lightbox.pswp);
	if (lightbox.pswp.currSlide.data._video != undefined) {
		lightbox.pswp.currSlide.data._video.close();
	}
});

lightbox.init();

Array.from(document.getElementsByClassName("text-spoiler")).forEach((spoiler) => {
	const button = spoiler.getElementsByClassName("button")[0];

	if (button != undefined) {
		function update() {
			if (spoiler.open) {
				button.textContent = "Show less";
			} else {
				button.textContent = "Show more";
			}
		}
		update();

		spoiler.addEventListener("toggle", update);
	}
});

Array.from(document.getElementsByClassName("plyr-video")).forEach((video) => {
	let player = new Plyr(video, {
		title: video.title,
		settings: ["loop"],
		disableContextMenu: false,
		hideControls: false,
		tooltips: { contrors: true, seek: true },
		listeners: {
			fullscreen: () => {
				lightbox.loadAndOpen(parseInt(video.dataset.pswpIndex), {
					gallery: video.closest(".photoswipe-gallery")
				});

				return false;
			}
		}
	});

	player.elements.container.title = video.title;
	video._plyrContainer = player.elements.container;
});

Array.from(document.getElementsByClassName("video-spoiler")).forEach((spoiler) => {
	const video = spoiler.getElementsByClassName("plyr-video")[0];

	if (video != undefined) {
		function update() {
			if (!spoiler.open) {
				video.pause();
			}
		}

		spoiler.addEventListener("toggle", update);
	}
});