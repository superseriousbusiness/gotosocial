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

import React, { StrictMode, useMemo } from "react";
import "./style.css";

import { createRoot } from "react-dom/client";
import { Provider } from "react-redux";
import { PersistGate } from "redux-persist/integration/react";
import { store, persistor } from "./redux/store";
import { Authorization } from "./components/authorization";
import Loading from "./components/loading";
import { Account } from "./lib/types/account";
import { BaseUrlContext, RoleContext, InstanceDebugContext } from "./lib/navigation/util";
import { SidebarMenu } from "./lib/navigation/menu";
import { Redirect, Route, Router } from "wouter";
import AdminMenu from "./views/admin/menu";
import ModerationMenu from "./views/moderation/menu";
import UserMenu from "./views/user/menu";
import UserRouter from "./views/user/router";
import { ErrorBoundary } from "./lib/navigation/error";
import ModerationRouter from "./views/moderation/router";
import AdminRouter from "./views/admin/router";
import { useInstanceV1Query } from "./lib/query/gts-api";

interface AppProps {
	account: Account;
}

export function App({ account }: AppProps) {
	const roles: string[] = useMemo(() => [ account.role.name ], [account]);
	const { data: instance } = useInstanceV1Query();
	
	return (
		<RoleContext.Provider value={roles}>
			<InstanceDebugContext.Provider value={instance?.debug ?? false}>
				<BaseUrlContext.Provider value={"/settings"}>
					<SidebarMenu>
						<UserMenu />
						<ModerationMenu />
						<AdminMenu />
					</SidebarMenu>
					<section className="with-sidebar">
						<Router base="/settings">
							<ErrorBoundary>
								<UserRouter />
								<ModerationRouter />
								<AdminRouter />
								{/*
								Ensure user ends up somewhere
								if they just open /settings.
							*/}
								<Route path="/"><Redirect to="/user/profile" /></Route>
							</ErrorBoundary>
						</Router>
					</section>
				</BaseUrlContext.Provider>
			</InstanceDebugContext.Provider>
		</RoleContext.Provider>
	);
}

function Main() {
	return (
		<Provider store={store}>
			<PersistGate
				loading={<section><Loading /></section>}
				persistor={persistor}
			>
				<Authorization App={App} />
			</PersistGate>
		</Provider>
	);
}

const root = createRoot(document.getElementById("root") as HTMLElement);
root.render(<StrictMode><Main /></StrictMode>);
