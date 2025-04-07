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

import React, { ReactNode, useEffect, useMemo, useState } from "react";
import { TextInput } from "../../../components/form/inputs";
import MutationButton from "../../../components/form/mutation-button";
import useFormSubmit from "../../../lib/form/submit";
import {
	useTwoFactorQRCodeURIMutation,
	useTwoFactorDisableMutation,
	useTwoFactorEnableMutation,
	useTwoFactorQRCodePngMutation,
} from "../../../lib/query/user/twofactor";
import { useTextInput } from "../../../lib/form";
import Loading from "../../../components/loading";
import { Error } from "../../../components/error";
import { HighlightedCode } from "../../../components/highlightedcode";
import { useDispatch } from "react-redux";
import { gtsApi } from "../../../lib/query/gts-api";

interface TwoFactorProps {
	twoFactorEnabledAt?: string,
	oidcEnabled?: boolean,
}

export default function TwoFactor({ twoFactorEnabledAt, oidcEnabled }: TwoFactorProps) {
	switch (true) {
		case oidcEnabled:
			// Can't enable if OIDC is in place.
			return <CannotEnable />;
		case twoFactorEnabledAt !== undefined:
			// Already enabled. Show the disable form.
			return <DisableForm twoFactorEnabledAt={twoFactorEnabledAt as string} />;	
		default:
			// Not enabled. Show the enable form.
			return <EnableForm />;
	}
}

function CannotEnable() {
	return (
		<form>
			<TwoFactorHeader
				blurb={
					<p>
						OIDC is enabled for your instance. To enable 2FA, you must use your
						instance's OIDC provider instead. Poke your admin for more information.
					</p>
				}
			/>
		</form>
	);
}

function EnableForm() {
	const form = { code: useTextInput("code") };
	const [ recoveryCodes, setRecoveryCodes ] = useState<string>();
	const dispatch = useDispatch();

	// Prepare trigger to submit the code and enable 2FA.
	// If the enable call is a success, set the recovery
	// codes state to a nice newline-separated text.
	const [submitForm, result] = useFormSubmit(form, useTwoFactorEnableMutation(), {
		changedOnly: true,
		onFinish: (res) => {
			const codes = res.data as string[];
			if (!codes) {
				return;
			}
			setRecoveryCodes(codes.join("\n"));
		},
	});

	// When the component is unmounted, clear the user
	// cache if 2FA was just enabled. This will prevent
	// the recovery codes from being shown again.
	useEffect(() => {
		return () => {
			if (recoveryCodes) {
				dispatch(gtsApi.util.invalidateTags(["User"]));
			}
		};
	}, [recoveryCodes, dispatch]);

	return (
		<form className="2fa-enable-form" onSubmit={submitForm}>
			<TwoFactorHeader
				blurb={
					<p>
						You can use this form to enable 2FA for your account.
						<br/>
						In your authenticator app, either scan the QR code, or copy
						the 2FA secret manually, and then enter a 2FA code to verify.
					</p>
				}
			/>
			{/*
				If the enable call was successful then recovery
				codes will now be set. Display these to the user.

				If the call hasn't been made yet, show the
				form to enable 2FA as normal.
			*/}
			{ recoveryCodes
				? <>
					<p>
						<b>Two-factor authentication is now enabled for your account!</b>
						<br/>From now on, you will need to provide a code from your authenticator app whenever you want to sign in.
						<br/>If you lose access to your authenticator app, you may also sign in by providing one of the below one-time recovery codes instead of a 2FA code.
						<br/>Once you have used a recovery code once, you will not be able to use it again!
						<br/><strong>You will not be shown these codes again, so copy them now into a safe place! Treat them like passwords!</strong>
					</p>
					<details>
						<summary>Show / hide codes</summary>
						<HighlightedCode
							code={recoveryCodes}
							lang="text"
						/>
					</details>
				</> 
				: <>
					<CodePng />
					<Secret />
					<TextInput
						name="code"
						field={form.code}
						label="2FA code from your authenticator app (6 numbers)"
						autoComplete="off"
						disabled={false}
						maxLength={6}
						minLength={6}
						pattern="^\d{6}$"
						readOnly={false}
					/>
					<MutationButton
						label="Enable 2FA"
						result={result}
						disabled={false}
					/>
				</>
			}
		</form>
	);
}

// Load and show QR code png only when
// the "Show QR Code" button is clicked.
function CodePng() {
	const [
		getPng, {
			isUninitialized,
			isLoading,
			isSuccess,
			data,
			error,
			reset,
		}
	] = useTwoFactorQRCodePngMutation();
	
	const [ content, setContent ] = useState<ReactNode>();
	useEffect(() => {
		if (isLoading) {
			setContent(<Loading />);
		} else if (isSuccess && data) {
			setContent(<img src={data} height="256" width="256" />);
		} else {
			setContent(<Error error={error} />);
		}
	}, [isLoading, isSuccess, data, error]);
	
	return (
		<>
			{ isUninitialized
				? <button
					disabled={false}
					onClick={(e) => {
						e.preventDefault();
						getPng();
					}}
				>Show QR Code</button>
				: <button
					disabled={false}
					onClick={(e) => {
						e.preventDefault();
						reset();
						setContent(null);
					}}
				>Hide QR Code</button>
			}
			{ content }
		</>
	);
}

// Get 2fa secret from server and
// load it into clipboard on click.
function Secret() {
	const [
		getURI,
		{
			isUninitialized,
			isSuccess,
			data,
			error,
			reset,
		},
	] = useTwoFactorQRCodeURIMutation();
	
	const [ buttonContents, setButtonContents ] = useState<ReactNode>();
	useEffect(() => {
		if (isUninitialized) {
			setButtonContents("Copy 2FA secret to clipboard");
		} else if (isSuccess && data) {
			const url = new URL(data);
			const secret = url.searchParams.get("secret");
			if (!secret) {
				throw "null secret";
			}
			navigator.clipboard.writeText(secret);
			setButtonContents("Copied!");
			setTimeout(() => { reset(); }, 3000);
		} else {
			setButtonContents(<Error error={error} />);
		}
	}, [isUninitialized, isSuccess, data, reset, error]);
	
	return (
		<button
			disabled={false}
			onClick={(e) => {
				e.preventDefault();
				getURI();
			}}
		>{buttonContents}</button>
	);
}

function DisableForm({ twoFactorEnabledAt }: { twoFactorEnabledAt: string }) {
	const enabledAt = useMemo(() => {
		const enabledAt = new Date(twoFactorEnabledAt);
		return <time dateTime={twoFactorEnabledAt}>{enabledAt.toDateString()}</time>;
	}, [twoFactorEnabledAt]);
	
	const form = {
		password: useTextInput("password"),
	};

	const [submitForm, result] = useFormSubmit(form, useTwoFactorDisableMutation());
	return (
		<form className="2fa-disable-form" onSubmit={submitForm}>
			<TwoFactorHeader
				blurb={
					<p>
						Two-factor auth is enabled for your account, since <b>{enabledAt}</b>.
						<br/>To disable 2FA, supply your password for verification and click "Disable 2FA".
					</p>
				}
			/>
			<TextInput
				type="password"
				name="password"
				field={form.password}
				label="Current password"
				autoComplete="current-password"
				disabled={false}
			/>
			<MutationButton
				label="Disable 2FA"
				result={result}
				disabled={false}
				className="danger"
			/>
		</form>
	);
}

function TwoFactorHeader({ blurb }: { blurb: ReactNode }) {
	return (
		<div className="form-section-docs">
			<h3>Two-Factor Authentication</h3>
			{blurb}
			<a
				href="https://docs.gotosocial.org/en/latest/user_guide/settings/#two-factor"
				target="_blank"
				className="docslink"
				rel="noreferrer"
			>
				Learn more about this (opens in a new tab)
			</a>
		</div>
	);
}
