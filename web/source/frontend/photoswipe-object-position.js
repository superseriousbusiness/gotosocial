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
	Code in this file adapted from:
	https://github.com/vovayatsyuk/photoswipe-object-position (MIT License).
*/

function getCroppedBoundsOffset(position, imageSize, thumbSize, zoomLevel) {
	const float = parseFloat(position); 
	return position.indexOf('%') > 0
		? (thumbSize - imageSize * zoomLevel) * float / 100
		: float;
}
 
function getCroppedZoomPan(position, min, max) {
	const float = parseFloat(position);
	return position.indexOf('%') > 0 ? min + (max - min) * float / 100 : float;
}

function getThumbnail(el) {
	return el.querySelector('img');
}

function getObjectPosition(el) {
	return getComputedStyle(el).getPropertyValue('object-position').split(' ');
}

export default class ObjectPosition {
	constructor(lightbox) {
		/**
		* Make pan adjustments if large image doesn't fit the viewport.
		*
		* Examples:
		* 1. When thumb object-position is 50% 0 (top part is initially visible)
		*    make sure you'll see the top part of the large image as well.
		* 2. When thumb object-position is 50% 100% (bottom part is initially visible)
		*    make sure you'll see the bottom part of the large image as well.
		*/
		lightbox.on('initialZoomPan', (event) => {
			const slide = event.slide;
			if (!slide.data.element) {
				// No thumbnail
				// image set.
				return;
			}
			
			const thumbnailImg = getThumbnail(slide.data.element);
			if (!thumbnailImg) {
				// No thumbnail
				// image set.
				return;
			}

			const [positionX, positionY] = getObjectPosition(thumbnailImg);

			if (positionX !== '50%' && slide.pan.x < 0) {
				slide.pan.x = getCroppedZoomPan(positionX, slide.bounds.min.x, slide.bounds.max.x);
			}
 
			if (positionY !== '50%' && slide.pan.y < 0) {
				slide.pan.y = getCroppedZoomPan(positionY, slide.bounds.min.y, slide.bounds.max.y);
			}
		});

		/**
		* Fix opening animation when thumb object-position is not 50% 50%.
		* https://github.com/dimsemenov/PhotoSwipe/pull/1868
		*/
		lightbox.addFilter('thumbBounds', (thumbBounds, itemData) => {
			if (!itemData.element) {
				// No thumbnail
				// image set.
				return;
			}
			
			const thumbnailImg = getThumbnail(itemData.element);
			if (!thumbnailImg) {
				// No thumbnail
				// image set.
				return;
			}

			const thumbAreaRect = thumbnailImg.getBoundingClientRect();
			const fillZoomLevel = thumbBounds.w / itemData.width;
			const [positionX, positionY] = getObjectPosition(thumbnailImg);

			if (positionX !== '50%') {
				const offsetX = getCroppedBoundsOffset(positionX, itemData.width, thumbAreaRect.width, fillZoomLevel);
				thumbBounds.x = thumbAreaRect.left + offsetX;
				thumbBounds.innerRect.x = offsetX;
			}

			if (positionY !== '50%') {
				const offsetY = getCroppedBoundsOffset(positionY, itemData.height, thumbAreaRect.height, fillZoomLevel);
				thumbBounds.y = thumbAreaRect.top + offsetY;
				thumbBounds.innerRect.y = offsetY;
			}

			return thumbBounds;
		});
	}
}
