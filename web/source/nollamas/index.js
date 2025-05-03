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

const explanation = "Your browser is currently solving a proof-of-work challenge designed to deter \"ai\" scrapers. This should take no more than a few seconds...";

document.addEventListener("DOMContentLoaded", function() {
	// Get the nollamas section container.
	const nollamas = document.querySelector(".nollamas");

	// Add some "loading" text to show that
	// a proof-of-work captcha is being done.
	const p = this.createElement("p");
	p.className = "nollamas-explanation";
	p.appendChild(document.createTextNode(explanation));
	nollamas.appendChild(p);

	// Add a loading spinner, but only
	// animate it if motion is allowed.
	const spinner = document.createElement("i");
	spinner.className = "fa fa-spinner fa-2x fa-fw nollamas-status";
	if (!window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
		spinner.className += " fa-pulse";
	}
	spinner.setAttribute("title","Solving...");
	spinner.setAttribute("aria-label", "Solving");
	nollamas.appendChild(spinner);

	// Read the challenge and difficulty from
	// data attributes on the nollamas section.
	const challenge = nollamas.dataset.nollamasChallenge;
	const difficulty = nollamas.dataset.nollamasDifficulty;

	console.log("challenge:", challenge);     // eslint-disable-line no-console
	console.log("difficulty:", difficulty); // eslint-disable-line no-console

	// Prepare the worker with task function.
	const worker = new Worker("/assets/dist/nollamasworker.js");
	const startTime = performance.now();
	worker.postMessage({
		challenge: challenge,
		difficulty: difficulty,
	});

	// Set the main worker function.
	worker.onmessage = function (e) {
		if (e.data.done) {
			const endTime = performance.now();
			const duration = endTime - startTime;

			// Remove spinner and replace it with a tick
			// and info about how long it took to solve.
			nollamas.removeChild(spinner);
			const solutionWrapper = document.createElement("div");
			solutionWrapper.className = "nollamas-status";

			const tick = document.createElement("i");
			tick.className = "fa fa-check fa-2x fa-fw";
			tick.setAttribute("title","Solved!");
			tick.setAttribute("aria-label", "Solved!");
			solutionWrapper.appendChild(tick);

			const took = document.createElement("span");
			const solvedText = `Solved after ${e.data.nonce} iterations, in ${duration.toString()}ms!`;
			took.appendChild(document.createTextNode(solvedText));
			solutionWrapper.appendChild(took);

			nollamas.appendChild(solutionWrapper);

			// Wait for 500ms before redirecting, to give
			// visitors a shot at reading the message, but
			// not so long that they have to wait unduly.
			setTimeout(() => {
				let url = new URL(window.location.href);
				url.searchParams.set("nollamas_solution", e.data.nonce);
				window.location.replace(url.toString());
			}, 500);
		}
	};
});
