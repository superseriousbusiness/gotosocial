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

const { combineReducers } = require("redux");
const { configureStore } = require("@reduxjs/toolkit");
const {
	persistStore,
	persistReducer,
	FLUSH,
	REHYDRATE,
	PAUSE,
	PERSIST,
	PURGE,
	REGISTER,
} = require("redux-persist");

const query = require("../lib/query/base");

const combinedReducers = combineReducers({
	oauth: require("./reducers/oauth").reducer,
	instances: require("./reducers/instances").reducer,
	temporary: require("./reducers/temporary").reducer,
	user: require("./reducers/user").reducer,
	admin: require("./reducers/admin").reducer,
	[query.reducerPath]: query.reducer
});

const persistedReducer = persistReducer({
	key: "gotosocial-settings",
	storage: require("redux-persist/lib/storage").default,
	stateReconciler: require("redux-persist/lib/stateReconciler/autoMergeLevel2").default,
	whitelist: ["oauth"],
}, combinedReducers);

const store = configureStore({
	reducer: persistedReducer,
	middleware: (getDefaultMiddleware) => {
		return getDefaultMiddleware({
			serializableCheck: {
				ignoredActions: [FLUSH, REHYDRATE, PAUSE, PERSIST, PURGE, REGISTER, "temporary/setScrollElement"]
			}
		}).concat(query.middleware);
	}
});

const persistor = persistStore(store);

module.exports = { store, persistor };