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

import { useMemo } from "react";

import { AdminAccount } from "../types/account";
import { store } from "../../redux/store";

import humanizeDuration from "humanize-duration";

export function yesOrNo(b: boolean): string {
	return b ? "yes" : "no";
}

export function UseOurInstanceAccount(account: AdminAccount): boolean {
	// Pull our own URL out of storage so we can
	// tell if account is our instance account.
	const ourDomain = useMemo(() => {
		const instanceUrlStr = store.getState().login.instanceUrl;
		if (!instanceUrlStr) {
			return "";
		}

		const instanceUrl = new URL(instanceUrlStr);
		return instanceUrl.host;
	}, []);

	return !account.domain && account.username == ourDomain;
}

/**
 * Uppercase first letter of given string.
 */
export function useCapitalize(i?: string): string {
	return useMemo(() => {
		if (i === undefined) {
			return "";
		}
		
		return i.charAt(0).toUpperCase() + i.slice(1); 
	}, [i]);
}

/**
 * Return human-readable string representation of given bytes.
 * 
 * Adapted from https://stackoverflow.com/a/14919494.
 */
export function useHumanReadableBytes(bytes: number): string {
	return useMemo(() => {
		const thresh = 1024;
		const digitPrecision = 2;
		const r = 10**digitPrecision;
		const units = ['KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB'];

		if (Math.abs(bytes) < thresh) {
			return bytes + ' B';
		}

		let u = -1;
		let threshed = bytes;
		do { threshed /= thresh; ++u;
		} while (Math.round(Math.abs(threshed) * r) / r >= thresh && u < units.length - 1);

		return threshed.toFixed(digitPrecision) + ' ' + units[u];
	}, [bytes]);
}

/**
 * Return human-readable string representation of given time in seconds.
 */
export function useHumanReadableDuration(seconds: number): string {
	return useMemo(() => {
		if (seconds % 2629746 === 0) {
			const n = seconds / 2629746;
			return n + " month" + (n !== 1 ? "s" : "");
		}
		
		const ms = seconds*1000;
		return humanizeDuration(ms);
	}, [seconds]);
}
