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

	This script is loaded in the document head, and deferred + async,
	so it's *usually* run after the user is already looking at the page.
	Put stuff in here that doesn't shift the layout, and it doesn't really
	matter whether it loads immediately. So, progressive enhancement stuff.
*/

const Photoswipe = require("photoswipe/dist/umd/photoswipe.umd.min.js");
const PhotoswipeLightbox = require("photoswipe/dist/umd/photoswipe-lightbox.umd.min.js");
const PhotoswipeCaptionPlugin = require("photoswipe-dynamic-caption-plugin").default;
const ObjectPosition = require("./photoswipe-object-position.js").default;
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

// Enable object-position plugin for lightbox so that css
// object-position property can be used on preview images.
new ObjectPosition(lightbox);

lightbox.addFilter('itemData', (item) => {
	const el = item.element;
	if (
		el &&
		el.classList.contains("plyr-video") &&
		el._plyrContainer !== undefined
	) {
		const parentNode = el._plyrContainer.parentNode;
		const loopingAuto = el.classList.contains("gifv");
		return {
			alt: el.getAttribute("alt"),
			_video: {
				open(c) {
					c.appendChild(el._plyrContainer);
					if (loopingAuto) {
						// Start playing
						// when opened.
						el._player.play();
					}
				},
				close() {
					parentNode.appendChild(el._plyrContainer);
				},
				pause() {
					el._player.pause();
				},
				play() {
					el._player.play();
				}
			},
			width: parseInt(el.dataset.pswpWidth),
			height: parseInt(el.dataset.pswpHeight),
			parentStatus: el.dataset.pswpParentStatus,
			attachmentId: el.dataset.pswpAttachmentId,
			loopingAuto: loopingAuto,
		};
	}
	return item;
});

// Open video when user moves to its slide.
lightbox.on("contentActivate", (e) => {
	const { content } = e;
	if (content.data._video != undefined) {
		content.data._video.open(content.element);
	}
});

// Pause + close video when user
// moves away from its slide.
lightbox.on("contentDeactivate", (e) => {
	const { content } = e;
	if (content.data._video != undefined) {
		content.data._video.pause();
		content.data._video.close();
	}
});

// Pause video when lightbox is closed.
lightbox.on("closingAnimationStart", function () {
	if (lightbox.pswp.currSlide.data._video != undefined) {
		lightbox.pswp.currSlide.data._video.close();
	}
});
lightbox.on("close", function () {
	if (lightbox.pswp.currSlide.data._video != undefined &&
		!lightbox.pswp.currSlide.data.loopingAuto) {
		lightbox.pswp.currSlide.data._video.pause();
	}
});

// Open video when lightbox is opened.
lightbox.on("openingAnimationEnd", function () {
	if (lightbox.pswp.currSlide.data._video != undefined) {
		lightbox.pswp.currSlide.data._video.play();
	}
});

// Add "open this post" link to lightbox UI.
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
				switch (true) {
					case pswp.currSlide.data.parentStatus !== undefined:
						// Link to parent status.
						el.href = pswp.currSlide.data.parentStatus;
						break;
					case pswp.currSlide.data.element !== undefined &&
						pswp.currSlide.data.element.dataset.pswpParentStatus !== undefined:
						// Link to parent status.
						el.href = pswp.currSlide.data.element.dataset.pswpParentStatus;
						break;
					default:
						// Link to profile.
						const location = window.location; 	
						el.href = "//" + location.host + location.pathname;
				}
			});
		}
	});
});

lightbox.init();

Array.from(document.getElementsByClassName("plyr-video")).forEach((video) => {
	const loopingAuto = !reduceMotion.matches && video.classList.contains("gifv");
	let player = new Plyr(video, {
		title: video.title,
		settings: [],
		// Only show controls for video and audio,
		// not looping soundless gifv. Don't show
		// volume slider as it's unusable anyway
		// when the video is inside a lightbox,
		// mute toggle will have to be enough.
		controls: loopingAuto
			? []
			: [
				'play-large',   // The large play button in the center
				'restart',      // Restart playback
				'rewind',       // Rewind by the seek time (default 10 seconds)
				'play',         // Play/pause playback
				'fast-forward', // Fast forward by the seek time (default 10 seconds)
				'current-time', // The current time of playback
				'duration',     // The full duration of the media
				'mute',         // Toggle mute
				'fullscreen',   // Toggle fullscreen
			],
		tooltips: { controls: true, seek: true },
		iconUrl: "/assets/plyr.svg",
		invertTime: false,
		hideControls: false,
		listeners: {
			play: (_) => {
				if (!inLightbox(video)) {
					// If the video isn't open in the lightbox
					// as the current photoswipe slide, clicking
					// on it to play it opens it in the lightbox.
					lightbox.loadAndOpen(parseInt(video.dataset.pswpIndex), {
						gallery: video.closest(".photoswipe-gallery")
					});
				} else if (!loopingAuto) {
					// If the video *is* open in the lightbox,
					// and it's not a looping gifv, clicking
					// play just plays or pauses the video.
					player.togglePlay();
				}
				return false;
			},
		}
	});

	player.elements.container.title = video.title;
	video._player = player;
	video._plyrContainer = player.elements.container;
});

// Return true if the photoswipe lightbox is
// open with this element as the current slide.
function inLightbox(element) {
	if (lightbox.pswp === undefined) {
		return false;
	}

	if (lightbox.pswp.currSlide === undefined) {
		return false;
	}

	return element.dataset.pswpAttachmentId ===
		lightbox.pswp.currSlide.data.attachmentId;
}

// When clicking anywhere that's not an open
// stats-info-more-content details dropdown,
// close that open dropdown.
document.body.addEventListener("click", (e) => {
	const openStats = document.querySelector("details.stats-more-info[open]");
	if (!openStats) {
		// No open stats
		// details element.
		return;
	}

	if (openStats.contains(e.target)) {
		// Click is within stats
		// element, leave it alone.
		return;
	}

	// Click was outside of
	// stats elements, close it. 
	openStats.removeAttribute("open");
});

// Scan for the first ListenBrainz profile field and replace
// its value with currently listening track if available.
//
// ListenBrainz allows a lot of leeway in usernames so be gentle here:
//
// See:
//
// - https://github.com/metabrainz/musicbrainz-server/blob/master/lib/MusicBrainz/Server/Form/Utils.pm#L264-L288
// - https://regex101.com/r/k5ij9F/1
const listenbrainzRe = new RegExp(/^https:\/\/listenbrainz\.org\/user\/([^/]+)\/$/, "u");
let calledListenBrainz = false;
document.querySelectorAll("div#profile-fields dl div.field").forEach((field) => {
	// If we called ListenBrainz once
	// already this page load, bail.
	if (calledListenBrainz) {
		return;
	}

	const k = field.querySelector("dt");
	if (!k) {
		// No <dt> inside this
		// field? Weird but OK.
		return;
	}

	const kText = k.textContent;
	if (kText === null) {
		// Also strange but
		// let's just bail.
		return;
	}
	
	// Check if key == "ListenBrainz" (case insensitive).
	if (kText.localeCompare("ListenBrainz", undefined, { sensitivity: "base" }) !== 0) {
		// Not interested.
		return;
	}

	// Get the value.
	const v = field.querySelector("dd");
	if (!v) {
		// No <dd> inside this
		// field? Weird but OK.
		return;
	}

	// Look for an <a> tag inside the <dd>.
	const oldAs = v.getElementsByTagName("a");
	if (oldAs.length !== 1) {
		// Nothing
		// in here.
		return;
	}

	const oldA = oldAs[0];
	const profileURL = oldA.textContent;
	if (!profileURL) {
		// Also strange but
		// let's just bail.
		return;
	}

	// We're looking for a listenbrainz URL.
	const match = profileURL.match(listenbrainzRe);
	if (match.length !== 2) {
		// Not a match.
		return;
	}
	const lbUsername = match[1];
	
	try {
		// MusicBrainz/ListenBrainz is very permissive
		// re: usernames so make sure to encode the URI
		// when doing the fetch, to avoid any shenanigans.
		const apiURL = encodeURI(`https://api.listenbrainz.org/1/user/${lbUsername}/playing-now`);
		fetch(apiURL).then(res => {
			// Mark that we
			// called LB already.
			calledListenBrainz = true;
			
			// Check result...
			if (!res.ok) {
				throw new Error(`Response status: ${res.status}`);
			}
			
			return res.json();
		}).then(json => {
			// Parse out the object.
			const payload = json.payload;
			if (!payload) {
				// Can't do anything
				// with no payload.
				return;
			}

			const listens = payload.listens;
			if (!listens || !Array.isArray(listens) || listens.length !== 1) {
				// Can't do anything
				// with no listens.
				return;
			}

			const listen = listens[0];
			const trackMetadata = listen.track_metadata;
			if (!trackMetadata) {
				// Can't do anything
				// with no track metadata.
				return;
			}

			const artistName = trackMetadata.artist_name;
			const trackName = trackMetadata.track_name;
			if (artistName === undefined || trackName === undefined) {
				// Can't display
				// this track.
				return;
			}

			// We can work with this.
			//
			// Rewrite the existing <dd> with the
			// current listening song, and keep the
			// link to the user's ListenBrainz profile.
			const vNew = document.createElement("dd");
			
			// Lil music note icon.
			const i = document.createElement("i");
			i.ariaHidden = "true";
			i.className = "fa fa-fw fa-music";
			vNew.appendChild(i);
			
			vNew.appendChild(document.createTextNode(" Now listening to: "));
			vNew.appendChild(document.createElement("br"));

			// Build the new link, taking
			// the href from the old link.
			const a = document.createElement("a");
			a.href = oldA.href;
			a.rel = "nofollow noreferrer noopener";
			a.target = "_blank";

			// Add track name in bold.
			const trackNameE = document.createElement("b");
			trackNameE.textContent = trackName;
			a.appendChild(trackNameE);

			// Add joiner in normal font.
			a.appendChild(document.createTextNode(" by "));
			
			// Add artist name in bold.
			const artistNameE = document.createElement("b");
			artistNameE.textContent = artistName;
			a.appendChild(artistNameE);

			// Put the link
			// in the definish.
			vNew.appendChild(a);

			// Do the replacement.
			field.replaceChild(vNew, v);
		});
	} catch (error) {
		// eslint-disable-next-line no-console
		console.error(error.message);
	}
});
