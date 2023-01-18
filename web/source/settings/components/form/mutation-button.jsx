/*
	GoToSocial
	Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

const React = require("react");
const { Error } = require("../error");

module.exports = function MutationButton({ label, result, disabled, showError = true, className = "", ...inputProps }) {
	let iconClass = "";
	const targetsThisButton = result.action == inputProps.name; // can also both be undefined, which is correct

	if (targetsThisButton) {
		if (result.isLoading) {
			iconClass = "fa-spin fa-refresh";
		} else if (result.isSuccess) {
			iconClass = "fa-check fadeout";
		}
	}

	return (<div>
		{(showError && targetsThisButton && result.error) &&
			<Error error={result.error} />
		}
		<button type="submit" className={"with-icon " + className} disabled={result.isLoading || disabled}	{...inputProps}>
			<i className={`fa fa-fw ${iconClass}`} aria-hidden="true"></i>
			{(targetsThisButton && result.isLoading)
				? "Processing..."
				: label
			}
		</button>
	</div>
	);
};