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

import React from "react";
import { Error } from "../error";

export interface MutationButtonProps extends React.DetailedHTMLProps<React.ButtonHTMLAttributes<HTMLButtonElement>, HTMLButtonElement> {
	label: string,
	result,
	disabled: boolean,
	showError?: boolean,
	className?: string,
	wrapperClassName?: string,
	submit?: boolean,
}

export default function MutationButton({
	label,
	result,
	disabled,
	showError = true,
	className = "",
	wrapperClassName = "",
	submit = true,
	...inputProps
}: MutationButtonProps) {
	let iconClass = "";
	// Can also both be undefined, which is correct.
	const targetsThisButton = result.action == inputProps.name; 

	if (targetsThisButton) {
		if (result.isLoading) {
			iconClass = " fa-spin fa-refresh";
		} else if (result.isSuccess) {
			iconClass = " fa-check fadeout";
		}
	}

	return (
		<div className={wrapperClassName ? wrapperClassName : "mutation-button"}>
			{(showError && targetsThisButton && result.error) &&
				<Error error={result.error} reset={result.reset} />
			}
			<button
				type={submit ? "submit" : "button"}
				className={"with-icon " + className}
				disabled={result.isLoading || disabled}
				{...inputProps}
			>
				<i className={`fa fa-fw${iconClass}`} aria-hidden="true"></i>
				{(targetsThisButton && result.isLoading)
					? "Processing..."
					: label
				}
			</button>
		</div>
	);
}
