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
	const seed = nollamas.dataset.nollamasSeed;
	const challenge = nollamas.dataset.nollamasChallenge;
	const threads = navigator.hardwareConcurrency;
	if (typeof(threads) != "number" || threads < 1) { threads = 1; }

	console.log("seed:", seed);           // eslint-disable-line no-console
	console.log("challenge:", challenge); // eslint-disable-line no-console
	console.log("threads:", threads);     // eslint-disable-line no-console

	// Create an array to track
	// all workers such that we
	// can terminate them all.
	const workers = [];
	const terminateAll = () => { workers.forEach((worker) => worker.terminate() ); };

	// Get time before task completion.
	const startTime = performance.now();

	// Prepare and start each worker,
	// adding them to the workers array.
	for (let i = 0; i < threads; i++) {
		const worker = new Worker("/assets/dist/nollamasworker.js");
		workers.push(worker);

		// On any error terminate.
		worker.onerror = (ev) => {
			console.error("worker error:", ev); // eslint-disable-line no-console
			terminateAll();
		};

		// Post worker message, where each
		// worker will compute nonces in range:
		// $threadNumber + $totalThreads * n
		worker.postMessage({
			challenge: challenge,
			threads: threads,
			thread: i,
			seed: seed,
		});

		// Set main on-success function.
		worker.onmessage = function (e) {
			if (e.data.done) {
				// Stop workers.
				terminateAll();

				// Log which thread found the solution.
				console.log("solution from:", e.data.thread); // eslint-disable-line no-console

				// Get total computation duration.
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
				const solvedText = `Solved after ${e.data.nonce} iterations by worker ${e.data.thread} of ${threads}, in ${duration.toString()}ms!`;
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
	}
});
