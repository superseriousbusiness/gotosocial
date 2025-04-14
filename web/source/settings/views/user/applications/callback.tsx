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
import { useSearch } from "wouter";
import { Error as ErrorCmp } from "../../../components/error";
import { useGetAccessTokenForAppMutation, useGetAppQuery } from "../../../lib/query/user/applications";
import { useCallbackURL } from "./common";
import useFormSubmit from "../../../lib/form/submit";
import { useValue } from "../../../lib/form";
import MutationButton from "../../../components/form/mutation-button";
import FormWithData from "../../../lib/form/form-with-data";
import { App } from "../../../lib/types/application";
import { OAuthAccessToken } from "../../../lib/types/oauth";

export function AppTokenCallback({}) {		
	// Read the callback authorization
	// information from the search params. 
	const search = useSearch();
	const urlQueryParams = new URLSearchParams(search);
	const code = urlQueryParams.get("code");
	const appId = urlQueryParams.get("state");
	const error = urlQueryParams.get("error");
	const errorDescription = urlQueryParams.get("error_description");

	if (error) {
		let errString = error;
		if (errorDescription) {
			errString += ": " + errorDescription;
		}
		if (error === "invalid_scope") {
			errString += ". You probably requested a token (sub-)scope that wasn't contained in the scopes of your application.";
		}
		const err = Error(errString);
		return <ErrorCmp error={err} />;
	}

	if (!code || !appId) {
		const err = Error("code or app id not defined");
		return <ErrorCmp error={err} />;
	}

	return(
		<>
			<FormWithData
				dataQuery={useGetAppQuery}
				queryArg={appId}
				DataForm={AccessForAppForm}
				{...{ code: code }}
			/>
		</>
	);
}


function AccessForAppForm({ data: app, code }: { data: App, code: string }) {
	const redirectURI = useCallbackURL();
	
	// Prepare to call /oauth/token to
	// exchange code for access token.
	const form = {
		client_id: useValue("client_id", app.client_id),
		client_secret: useValue("client_secret", app.client_secret),
		redirect_uri: useValue("redirect_uri", redirectURI),
		code: useValue("code", code),
		grant_type: useValue("grant_type", "authorization_code"),
		
	};
	const [ submit, result ] = useFormSubmit(form, useGetAccessTokenForAppMutation());

	return (
		<form
			className="access-token-receive-form"	
			onSubmit={submit}
		>
			<div className="form-section-docs">
				<h2>Receive Access Token</h2>
				<p>
					To receive your user-level access token for application<b>{app.name}</b>, click on the button below.
					<br/>Your access token will be shown once and only once.
					<br/><strong>Your access token provides access to your account; store it as carefully as you would store a password!</strong>
				</p>
				<a
					href="https://docs.gotosocial.org/en/latest/api/authentication/#verifying"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about how to use your access token (opens in a new tab)
				</a>
			</div>
			
			{ result.data
				? <div className="access-token-frame monospace">{(result.data as OAuthAccessToken).access_token}</div>
				: <div className="access-token-frame closed"><i className="fa fa-eye-slash" aria-hidden={true}></i></div>
			}
			
			<MutationButton
				label="I understand, show me the token!"
				result={result}
				disabled={result.data || result.isError}
			/>
		</form>
	);
}
