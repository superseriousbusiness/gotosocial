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

let compute = async function(challengeStr, diffStr) {
	const textEncoder = new TextEncoder();

	// Get difficulty1 as number and generate
	// expected zero ASCII prefix to check for.
	const diff1 = parseInt(diffStr, 10);
	const zeros = "0".repeat(diff1);

	// Calculate hex encoded prefix required to check solution, where we
	// need diff1 no. chars in hex, and hex encoding doubles input length.
	const prefixLen = diff1 / 2 + (diff1 % 2 != 0 ? 2 : 0);

	let nonce = 0;
	while (true) { // eslint-disable-line no-constant-condition

		// Create possible solution string from challenge string + nonce.
		const solution = textEncoder.encode(challengeStr + nonce.toString());

		// Generate SHA256 hashsum of solution string, and hex encode the
		// necessary prefix length we need to check for a valid solution.
		const prefixArray = Array.from(sha256(solution).slice(0, prefixLen));
		const prefixHex = prefixArray.map(b => b.toString(16).padStart(2, "0")).join("");

		// Check if the hex encoded hash has
		// difficulty defined zeroes prefix.
		if (prefixHex.startsWith(zeros)) {
			return nonce;
		}

		// Iter.
		nonce++;
	}
};

onmessage = async function(e) {
	console.log('worker started'); // eslint-disable-line no-console

	const challenge = e.data.challenge;
	const difficulty = e.data.difficulty;

	// Compute the nonce that produces solution with args.
	let nonce = await compute(challenge, difficulty);

	// Post the solution nonce back to caller.
	postMessage({ nonce: nonce, done: true });
};
