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

onmessage = async function(e) {
	console.log('worker started'); // eslint-disable-line no-console

	const challenge = e.data.challenge;
	const textEncoder = new TextEncoder();

	// Get difficulty and generate the expected
	// zero ASCII prefix to check for in hashes.
	const difficultyStr = e.data.difficulty;
	const difficulty = parseInt(difficultyStr, 10);
	const zeroPrefix = '0'.repeat(difficulty);

	let nonce = 0;
	while (true) { // eslint-disable-line no-constant-condition

		// Create possible solution string from challenge + nonce.
		const solution = textEncoder.encode(challenge + nonce.toString());

		// Generate SHA256 hashsum of solution string and hex encode the result.
		const hashBuffer = await crypto.subtle.digest('SHA-256', solution);
		const hashArray = Array.from(new Uint8Array(hashBuffer));
		const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');

		// Check if the hex encoded hash has
		// difficulty defined zeroes prefix.
		if (hashHex.startsWith(zeroPrefix)) {
			postMessage({ nonce: nonce, done: true });
			break;
		}

		// Iter.
		nonce++;
	}
};
