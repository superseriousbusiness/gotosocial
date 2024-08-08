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

const Photoswipe = require("photoswipe/dist/umd/photoswipe.umd.min.js");
const PhotoswipeLightbox = require("photoswipe/dist/umd/photoswipe-lightbox.umd.min.js");
const PhotoswipeCaptionPlugin = require("photoswipe-dynamic-caption-plugin").default;
const Plyr = require("plyr");
const Prism = require("./prism.js");

Prism.manual = true;
Prism.highlightAll();

const reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)');

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
	if (
		el &&
		el.classList.contains("plyr-video") &&
		el._plyrContainer !== undefined
	) {
		const parentNode = el._plyrContainer.parentNode;
		return {
			alt: el.getAttribute("alt"),
			_video: {
				open(c) {
					c.appendChild(el._plyrContainer);
				},
				close() {
					parentNode.appendChild(el._plyrContainer);
				},
				pause() {
					el._player.pause();
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
		content.data._video.open(content.element);
	}
});

lightbox.on("contentDeactivate", (e) => {
	const { content } = e;
	if (content.data._video != undefined) {
		content.data._video.pause();
		content.data._video.close();
	}
});

lightbox.on("close", function () {
	if (lightbox.pswp.currSlide.data._video != undefined) {
		lightbox.pswp.currSlide.data._video.close();
	}
});

lightbox.init();

function dynamicSpoiler(className, updateFunc) {
	Array.from(document.getElementsByClassName(className)).forEach((spoiler) => {
		const update = updateFunc(spoiler);
		if (update) {
			update();
			spoiler.addEventListener("toggle", update);
		}
	});
}

dynamicSpoiler("text-spoiler", (spoiler) => {
	const button = spoiler.querySelector(".button");

	return () => {
		button.textContent = spoiler.open
			? "Show less"
			: "Show more";
	};
});

dynamicSpoiler("media-spoiler", (spoiler) => {
	const eye = spoiler.querySelector(".eye.button");
	const video = spoiler.querySelector(".plyr-video");
	const loopingAuto = !reduceMotion.matches && video != null && video.classList.contains("gifv");

	return () => {
		if (spoiler.open) {
			eye.setAttribute("aria-label", "Hide media");
		} else {
			eye.setAttribute("aria-label", "Show media");
			if (video && !loopingAuto) {
				video.pause();
			}
		}
	};
});

Array.from(document.getElementsByClassName("plyr-video")).forEach((video) => {
	const loopingAuto = !reduceMotion.matches && video.classList.contains("gifv");
	
	if (loopingAuto) {
		// If we're able to play this as a
		// looping gifv, then do so, else fall
		// back to user-controllable video player.
		video.draggable = false;
		video.autoplay = true;
		video.loop = true;
		video.classList.remove("photoswipe-slide");
		video.classList.remove("plry-video");
		video.load();
		video.play();
		return;
	}
	
	let player = new Plyr(video, {
		title: video.title,
		settings: ["loop"],
		disableContextMenu: false,
		hideControls: false,
		tooltips: { contrors: true, seek: true },
		iconUrl: "/assets/plyr.svg",
		listeners: {
			fullscreen: () => {
				if (player.playing) {
					setTimeout(() => {
						player.play();
					}, 1);
				}
				lightbox.loadAndOpen(parseInt(video.dataset.pswpIndex), {
					gallery: video.closest(".photoswipe-gallery")
				});

				return false;
			}
		}
	});

	player.elements.container.title = video.title;
	video._player = player;
	video._plyrContainer = player.elements.container;
});
