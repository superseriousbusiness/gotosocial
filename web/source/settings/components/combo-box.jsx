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

const React = require("react");

const {
	Combobox,
	ComboboxItem,
	ComboboxPopover,
} = require("ariakit/combobox");

module.exports = function ComboBox({state, items, label, placeHolder, children}) {
	return (
		<div className="form-field combobox-wrapper">
			<label>
				{label}
				<div className="row">
					<Combobox
						state={state}
						placeholder={placeHolder}
						className="combobox input"
					/>
					{children}
				</div>
			</label>
			<ComboboxPopover state={state} className="popover">
				{items.map(([key, value]) => (
					<ComboboxItem className="combobox-item" key={key} value={key}>
						{value}
					</ComboboxItem>
				))}
			</ComboboxPopover>
		</div>
	);
};