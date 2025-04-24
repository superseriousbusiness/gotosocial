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

document.addEventListener('DOMContentLoaded', function() {
	// Read the challenge and difficulty from
	// data attributes on the nollamas section.
	const nollamas = document.querySelector(".nollamas");
	const challenge = nollamas.dataset.nollamasChallenge;
	const difficulty = nollamas.dataset.nollamasDifficulty;

	console.log('challenge:', challenge);   // eslint-disable-line no-console
	console.log('difficulty:', difficulty); // eslint-disable-line no-console

	// Prepare the worker with task function.
	const worker = new Worker("/assets/dist/nollamasworker.js");
	worker.postMessage({
		challenge: challenge,
		difficulty: difficulty,
	});

	// Set the main worker function.
	worker.onmessage = function (e) {
		if (e.data.done) {
			console.log('solution found for:', e.data.nonce); // eslint-disable-line no-console
			let url = new URL(window.location.href);
			url.searchParams.set('nollamas_solution', e.data.nonce);
			window.location.href = url.toString();
		}
	};
});
