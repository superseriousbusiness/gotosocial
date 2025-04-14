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

import React, { useEffect, useRef } from "react";
import { MediaAttachment } from "../lib/types/status";
import { decode } from "blurhash";

export default function BlurhashCanvas({ media }: { media: MediaAttachment }) {
	const hash = media.blurhash;
	const thumbHeight = media.meta.small.height;
	const thumbWidth = media.meta.small.width;
	const thumbAspect = media.meta.small.aspect;	

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

	const canvasRef = useRef<HTMLCanvasElement>(null);

	useEffect(() => {
		const ctx = canvasRef.current?.getContext("2d");
		if (!ctx) {
			return;
		}
    
		const imageData = ctx.createImageData(
			useWidth,
			useHeight,
		);
		imageData.data.set(pixels);
		ctx.putImageData(imageData, 0, 0);
	}, [pixels, useHeight, useWidth]);

	return (
		<canvas
			width={useWidth}
			height={useHeight}
			ref={canvasRef}
		></canvas>
	);
}
