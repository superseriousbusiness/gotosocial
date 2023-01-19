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

module.exports = function CheckList({ field, Component, header = " All", ...componentProps }) {
	return (
		<div className="checkbox-list list">
			<label className="header">
				<input
					ref={field.toggleAll.ref}
					type="checkbox"
					onChange={field.toggleAll.onChange}
					checked={field.toggleAll.value === 1}
				/> {header}
			</label>
			{Object.values(field.value).map((entry) => (
				<CheckListEntry
					key={entry.key}
					onChange={(value) => field.onChange(entry.key, value)}
					entry={entry}
					Component={Component}
					componentProps={componentProps}
				/>
			))}
		</div>
	);
};

function CheckListEntry({ entry, onChange, Component, componentProps }) {
	return (
		<label className="entry">
			<input
				type="checkbox"
				onChange={(e) => onChange({ checked: e.target.checked })}
				checked={entry.checked}
			/>
			<Component entry={entry} onChange={onChange} {...componentProps} />
		</label>
	);
}