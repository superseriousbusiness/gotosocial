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

import React, { useState } from "react";
import { useLocation, useParams } from "wouter";
import FormWithData from "../../../lib/form/form-with-data";
import BackButton from "../../../components/back-button";
import { useBaseUrl } from "../../../lib/navigation/util";
import { useDeleteAppMutation, useGetAppQuery, useGetOOBAuthCodeMutation } from "../../../lib/query/user/applications";
import { App } from "../../../lib/types/application";
import { useAppWebsite, useCallbackURL, useCreated, useRedirectURIs } from "./common";
import MutationButton from "../../../components/form/mutation-button";
import { useTextInput } from "../../../lib/form";
import { TextInput } from "../../../components/form/inputs";
import { useScopesPermittedBy, useScopesValidator } from "../../../lib/util/formvalidators";

export default function AppDetail({ }) {
	const params: { appId: string } = useParams();
	const baseUrl = useBaseUrl();
	const backLocation: String = history.state?.backLocation ?? `~${baseUrl}`;

	return (
		<div className="application-details">
			<h1><BackButton to={backLocation}/> Application Details</h1>
			<FormWithData
				dataQuery={useGetAppQuery}
				queryArg={params.appId}
				DataForm={AppDetailForm}
				{...{ backLocation: backLocation }}
			/>
		</div>
	);
}

function AppDetailForm({ data: app, backLocation }: { data: App, backLocation: string }) {	
	return (
		<>
			<AppBasicInfo app={app} />
			<AccessTokenForm app={app} />
			<DeleteAppForm app={app} backLocation={backLocation} />
		</>
	);
}

function AppBasicInfo({ app }: { app: App }) {
	const appWebsite = useAppWebsite(app);
	const created = useCreated(app);
	const redirectURIs = useRedirectURIs(app);
	const [ showClient, setShowClient ] = useState(false);
	const [ showSecret, setShowSecret ] = useState(false);

	return (
		<dl className="info-list">
			<div className="info-list-entry">
				<dt>Name:</dt>
				<dd className="text-cutoff">{app.name}</dd>
			</div>

			{ appWebsite && 
			<div className="info-list-entry">
				<dt>Website:</dt>
				<dd>{appWebsite}</dd>
			</div>
			}

			<div className="info-list-entry">
				<dt>Created:</dt>
				<dd>{created}</dd>
			</div>

			<div className="info-list-entry">
				<dt>Scopes:</dt>
				<dd className="monospace">{app.scopes.join(" ")}</dd>
			</div>

			<div className="info-list-entry">
				<dt>Redirect URI(s):</dt>
				<dd className="monospace">{redirectURIs}</dd>
			</div>

			<div className="info-list-entry">
				<dt>Vapid key:</dt>
				<dd className="monospace">{app.vapid_key}</dd>
			</div>

			<div className="info-list-entry">
				<dt>Client ID:</dt>
				{ showClient
					? <dd className="monospace">{app.client_id}</dd>
					: <dd><button onClick={() => setShowClient(true)}>Show client ID</button></dd>
			 	}
			</div>

			<div className="info-list-entry">
				<dt>Client secret:</dt>
				{ showSecret
					? <dd className="monospace">{app.client_secret}</dd>
					: <dd><button onClick={() => setShowSecret(true)}>Show secret</button></dd>
			 	}
			</div>
		</dl>
	);
}

function AccessTokenForm({ app }: { app: App }) {
	const [ getOOBAuthCode, result ] = useGetOOBAuthCodeMutation();
	const permittedScopes = useScopesPermittedBy();
	const validateScopes = useScopesValidator();
	const scope = useTextInput("scope", {
		defaultValue: app.scopes.join(" "),
		validator: (wantsScopesStr: string) => {
			if (wantsScopesStr === "") {
				return "";
			}

			// Check requested scopes are valid scopes.
			const wantsScopes = wantsScopesStr.split(" ");
			const invalidScopesMsg = validateScopes(wantsScopes);
			if (invalidScopesMsg !== "") {
				return invalidScopesMsg;
			}

			// Check requested scopes are permitted by the app.
			return permittedScopes(app.scopes, wantsScopes);
		}
	});
	
	const callbackURL = useCallbackURL();
	const disabled = !app.redirect_uris.includes(callbackURL);
	return (
		<form
			autoComplete="off"
			onSubmit={(e) => {
				e.preventDefault();
				getOOBAuthCode({
					app,
					scope: scope.value ?? "",
					redirectURI: callbackURL,
				});
			}}
		>
			<div className="form-section-docs">
				<h2>Request An API Access Token</h2>
				<p>
					If your application redirect URIs includes the settings panel callback URL,
					you can use this section to request an access token that you can use to make API calls.
					<br/>The token scopes specified below must be equal to, or a subset of, the scopes
					you provided when you created the application.
					<br/>After clicking "Request access token", you will be redirected to the sign in
					page for your instance, where you must provide your credentials in order to authorize
					your application to act on your behalf. You will then be redirected again to a page
					where you can view your new access token.
				</p>
				<a
					href="https://docs.gotosocial.org/en/latest/api/authentication/"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about the OAuth authentication flow (opens in a new tab)
				</a>
			</div>

			<TextInput
				className="monospace"
				field={scope}
				label="Token scopes (space-separated list)"
				autoCapitalize="off"
				autoCorrect="off"
				disabled={disabled}
			/>

			<MutationButton
				disabled={disabled}
				label="Request access token"
				result={result}
			/>
		</form>
	);
}

function DeleteAppForm({ app, backLocation }: { app: App, backLocation: string }) {
	const [ _location, setLocation ] = useLocation();
	const [ deleteApp, result ] = useDeleteAppMutation();

	return (
		<form>
			<div className="form-section-docs">
				<h2>Delete Application</h2>
				<p>
					You can use this button to delete the application.
					<br/>Any tokens created by the application will also be deleted.
				</p>
			</div>
			<MutationButton
				label={`Delete`}
				title={`Delete`}
				type="button"
				className="button danger"
				onClick={(e) => {
					e.preventDefault();
					deleteApp(app.id);
					setLocation(backLocation);
				}}
				disabled={false}
				showError={false}
				result={result}
			/>
		</form>
	);
}
