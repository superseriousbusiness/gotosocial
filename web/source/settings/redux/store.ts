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

import { combineReducers } from "redux";
import { configureStore } from "@reduxjs/toolkit";
import {
	persistStore,
	persistReducer,
	FLUSH,
	REHYDRATE,
	PAUSE,
	PERSIST,
	PURGE,
	REGISTER,
} from "redux-persist";

import { loginSlice } from "./login";
import { gtsApi } from "../lib/query/gts-api";

const combinedReducers = combineReducers({
	[gtsApi.reducerPath]: gtsApi.reducer,
	login: loginSlice.reducer,
});

const persistedReducer = persistReducer({
	key: "gotosocial-settings",
	storage: require("redux-persist/lib/storage").default,
	stateReconciler: require("redux-persist/lib/stateReconciler/autoMergeLevel1").default,
	whitelist: ["login"],
	migrate: async (state) => {
		if (state == undefined) {
			return state;
		}

		// This is a cheeky workaround for
		// redux-persist being a stickler.
		let anyState = state as any; 
		if (anyState?.login != undefined) {
			anyState.login.expectingRedirect = false;
		}

		return anyState;
	}
}, combinedReducers);

export const store = configureStore({
	reducer: persistedReducer,
	middleware: (getDefaultMiddleware) => {
		return getDefaultMiddleware({
			serializableCheck: {
				ignoredActions: [
					FLUSH,
					REHYDRATE,
					PAUSE,
					PERSIST,
					PURGE,
					REGISTER,
				],
				ignoredPaths: ['api.queries.twoFactorQRCodePng(undefined).data.data'],
			}
		}).concat(gtsApi.middleware);
	}
});

export const persistor = persistStore(store);

export type AppDispatch = typeof store.dispatch;
export type RootState = ReturnType<typeof store.getState>;
