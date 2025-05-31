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

import sha256 from "./sha256";

let compute = async function(seedStr, challengeStr, start, iter) {
	const textEncoder = new TextEncoder();

	let nonce = start;
	while (true) { // eslint-disable-line no-constant-condition

		// Create possible solution string from challenge string + nonce.
		const solution = textEncoder.encode(seedStr + nonce.toString());

		// Generate hex encoded SHA256 hashsum of solution.
		const hashArray = Array.from(sha256(solution));
		const hashAsHex = hashArray.map(b => b.toString(16).padStart(2, "0")).join("");

		// Check whether hex encoded
		// solution matches challenge.
		if (hashAsHex == challengeStr) {
			return nonce;
		}

		// Iter nonce.
		nonce += iter;
	}
};

onmessage = async function(e) {
	const thread = e.data.thread;
	const threads = e.data.threads;
	console.log("worker started:", thread); // eslint-disable-line no-console

	// Compute nonce value that produces 'challenge', for our allotted thread.
	let nonce = await compute(e.data.seed, e.data.challenge, thread, threads);

	// Post the solution nonce back to caller with thread no.
	postMessage({ nonce: nonce, done: true, thread: thread });
};
