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

import React, { PropsWithChildren } from "react";
import { Link, useRoute } from "wouter";
import {
	BaseUrlContext,
	MenuLevelContext,
	useBaseUrl,
	useHasPermission,
	useMenuLevel,
} from "./util";
import UserLogoutCard from "../../components/user-logout-card";
import { nanoid } from "nanoid";

export interface MenuItemProps {
	/**
	 * Name / title of this menu item.
	 */
	name?: string;

	/**
	 * Url path component for this menu item.
	 */
	itemUrl: string;

	/**
	 * If this menu item is a category containing
	 * children, which child should be selected by
	 * default when category title is clicked.
	 * 
	 * Optional, use for categories only.
	 */
	defaultChild?: string;

	/**
	 * Permissions required to access this
	 * menu item (none, "moderator", "admin").
	 */
	permissions?: string[];

	/**
	 * Fork-awesome string to render
	 * icon for this menu item.
	 */
	icon?: string;
}

export function MenuItem(props: PropsWithChildren<MenuItemProps>) {
	const {
		name,
		itemUrl,
		defaultChild,
		permissions,
		icon,
		children,
	} = props;
	
	// Derive where this item is
	// in terms of URL routing.
	const baseUrl = useBaseUrl();
	const thisUrl = [ baseUrl, itemUrl ].join('/');

	// Derive where this item is in
	// terms of nesting within the menu.
	const thisLevel = useMenuLevel();
	const nextLevel = thisLevel+1;
	const topLevel = thisLevel === 0;

	// Check whether this item is currently active
	// (ie., user has selected it in the menu).
	//
	// This uses a wildcard to mark both parent
	// and relevant child as active.
	//
	// See:
	// https://github.com/molefrog/wouter?tab=readme-ov-file#useroute-route-matching-and-parameters
	const [isActive] = useRoute([ thisUrl, "*?" ].join("/"));

	// Don't render item if logged-in user
	// doesn't have permissions to use it.
	if (!useHasPermission(permissions)) {
		return null;
	}

	// Check whether this item has children.
	const hasChildren = children !== undefined;
	const childrenArray = hasChildren && Array.isArray(children);

	// Class name of the item varies depending
	// on where it is in the menu, and whether
	// it has children beneath it or not.
	const classNames: string[] = [];
	if (topLevel) {
		classNames.push("category", "top-level");
	} else {
		switch (true) {
			case thisLevel === 1 && hasChildren:
				classNames.push("category", "expanding");
				break;
			case thisLevel === 1 && !hasChildren:
				classNames.push("view", "expanding");
				break;
			case thisLevel >= 2 && hasChildren:
				classNames.push("nested", "category");
				break;
			case thisLevel >= 2 && !hasChildren:
				classNames.push("nested", "view");
				break;
		}
	}

	if (isActive) {
		classNames.push("active");
	}

	let content: React.JSX.Element | null;
	if ((isActive || topLevel) && childrenArray) {
		// Render children as a nested list.
		content = <ul>{children}</ul>;
	} else if (isActive && hasChildren) {
		// Render child as solo element.
		content = <>{children}</>;
	} else {
		// Not active: hide children.
		content = null;
	}

	// If a default child is defined, this item should point to that.
	const href = defaultChild ? [ thisUrl, defaultChild ].join("/") : thisUrl;

	return (
		<li key={nanoid()} className={classNames.join(" ")}>
			<Link href={href} className="title">
				<span>
					{icon && <i className={`icon fa fa-fw ${icon}`} aria-hidden="true" />}
					{name}
				</span>
			</Link>
			{ content && 
				<BaseUrlContext.Provider value={thisUrl}>
					<MenuLevelContext.Provider value={nextLevel}>
						{content}
					</MenuLevelContext.Provider>
				</BaseUrlContext.Provider>
			}
		</li>
	);
}

export interface SidebarMenuProps{}

export function SidebarMenu({ children }: PropsWithChildren<SidebarMenuProps>) {
	return (
		<div className="sidebar">
			<UserLogoutCard />
			<nav className="menu-tree">
				<MenuLevelContext.Provider value={0}>
					<ul className="top-level">
						{children}
					</ul>
				</MenuLevelContext.Provider>
			</nav>
		</div>
	);
}
