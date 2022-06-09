/*
	GoToSocial
	Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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


// WARNING: currently dependencies get deduplicated with factor-bundle, but 
// our frontend templates don't load the common bundle.js since it contains React etc
// so we can't use any dependencies that would deduplicate with the other files

Array.from(document.getElementsByClassName("spoiler-label")).forEach((label) => {
	let checkbox = document.getElementById(label.htmlFor);
	console.log(label, checkbox);
	if (checkbox != undefined) {
		function update() {
			if(checkbox.checked) {
				label.innerHTML = "Show more";
			} else {
				label.innerHTML = "Show less";
			}
		}
		update();
	
		label.addEventListener("click", () => {setTimeout(update, 1);});
	}
});
