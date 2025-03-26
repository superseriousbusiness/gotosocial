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
	// Bit darker than default 0.8.
	bgOpacity: 0.9,
	loop: false,
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
			height: parseInt(el.dataset.pswpHeight),
			parentStatus: el.dataset.pswpParentStatus,
			attachmentId: el.dataset.pswpAttachmentId,
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

lightbox.on('uiRegister', function() {
	lightbox.pswp.ui.registerElement({
		name: 'open-post-link',
		ariaLabel: 'Open post',
		order: 8,
		isButton: true,
		tagName: "a",
		html: '<span title="Open post"><span class="sr-only">Open post</span><i class="fa fa-lg fa-external-link-square" aria-hidden="true"></i></span>',
		onInit: (el, pswp) => {
			el.setAttribute('target', '_blank');
			el.setAttribute('rel', 'noopener');
			pswp.on('change', () => {
				el.href = pswp.currSlide.data.parentStatus
					? pswp.currSlide.data.parentStatus
					: pswp.currSlide.data.element.dataset.pswpParentStatus;
			});
		  }
	});
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
		settings: [],
		controls: ['play-large', 'play', 'progress', 'current-time', 'volume', 'mute', 'fullscreen'],
		disableContextMenu: false,
		hideControls: false,
		tooltips: { controls: true, seek: true },
		iconUrl: "/assets/plyr.svg",
		invertTime: false,
		listeners: {
			fullscreen: () => {
				// Check if the photoswipe lightbox is
				// open with this as the current slide.
				const alreadyInLightbox = (
					lightbox.pswp !== undefined &&
					video.dataset.pswpAttachmentId === lightbox.pswp.currSlide.data.attachmentId
				);
				
				if (alreadyInLightbox) {
					// If this video is already open as the
					// current photoswipe slide, the fullscreen
					// button toggles proper fullscreen.
					player.fullscreen.toggle();
				} else {
					// Otherwise the fullscreen button opens
					// the video as current photoswipe slide.
					//
					// (Don't pause the video while it's
					// being transitioned to a slide.)
					if (player.playing) {
						setTimeout(() => player.play(), 1);
					}
					lightbox.loadAndOpen(parseInt(video.dataset.pswpIndex), {
						gallery: video.closest(".photoswipe-gallery")
					});
				}
				return false;
			}
		}
	});

	player.elements.container.title = video.title;
	video._player = player;
	video._plyrContainer = player.elements.container;
});

Array.from(document.getElementsByTagName('time')).forEach(timeTag => {
	const datetime = timeTag.getAttribute('datetime');
	const currentText = timeTag.textContent.trim();
	// Only format if current text contains precise time.
	if (currentText.match(/\d{2}:\d{2}/)) {
		const date = new Date(datetime);
		timeTag.textContent = date.toLocaleString(
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
	}
});
