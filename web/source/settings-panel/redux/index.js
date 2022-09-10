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

const { createStore, combineReducers, applyMiddleware } = require("redux");
const { persistStore, persistReducer } = require("redux-persist");
const thunk = require("redux-thunk").default;
const { composeWithDevTools } = require("redux-devtools-extension");

const persistConfig = {
	key: "gotosocial-settings",
	storage: require("redux-persist/lib/storage").default,
	stateReconciler: require("redux-persist/lib/stateReconciler/autoMergeLevel2").default,
	whitelist: ["oauth"],
	blacklist: ["temporary"]
};

const combinedReducers = combineReducers({
	oauth: require("./reducers/oauth").reducer,
	instances: require("./reducers/instances").reducer,
	temporary: require("./reducers/temporary").reducer,
});

const persistedReducer = persistReducer(persistConfig, combinedReducers);
const composedEnhancer = composeWithDevTools(applyMiddleware(thunk));

const store = createStore(persistedReducer, composedEnhancer);
const persistor = persistStore(store);

module.exports = { store, persistor };