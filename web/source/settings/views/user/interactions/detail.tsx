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

import React, { useMemo } from "react";
import { useLocation, useParams } from "wouter";
import FormWithData from "../../../lib/form/form-with-data";
import BackButton from "../../../components/back-button";
import { useBaseUrl } from "../../../lib/navigation/util";
import { useApproveInteractionRequestMutation, useGetInteractionRequestQuery, useRejectInteractionRequestMutation } from "../../../lib/query/user/interactions";
import { InteractionRequest } from "../../../lib/types/interaction";
import { useIcon, useNoun, useVerbed } from "./util";
import MutationButton from "../../../components/form/mutation-button";
import { Status } from "../../../components/status";

export default function InteractionRequestDetail({ }) {
	const params: { reqId: string } = useParams();
	const baseUrl = useBaseUrl();
	const backLocation: String = history.state?.backLocation ?? `~${baseUrl}`;

	return (
		<div className="interaction-request-detail">
			<h1><BackButton to={backLocation}/> Interaction Request Details</h1>
			<FormWithData
				dataQuery={useGetInteractionRequestQuery}
				queryArg={params.reqId}
				DataForm={InteractionRequestDetailForm}
				{...{ backLocation: backLocation }}
			/>
		</div>
	);
}

function InteractionRequestDetailForm({ data: req, backLocation }: { data: InteractionRequest, backLocation: string }) {	
	const [ _location, setLocation ] = useLocation();
	
	const [ approve, approveResult ] = useApproveInteractionRequestMutation();
	const [ reject, rejectResult ] = useRejectInteractionRequestMutation();
	
	const verbed = useVerbed(req.type);
	const noun = useNoun(req.type);
	const icon = useIcon(req.type);

	const strap = useMemo(() => {
		return "@" + req.account.acct + " " + verbed + " your post.";
	}, [req.account, verbed]);

	return (
		<>
			<span className="overview">
				<i
					className={`fa fa-fw ${icon}`}
					aria-hidden="true"
				/> <strong>{strap}</strong>
			</span>
			
			<h2>You wrote:</h2>
			<div className="thread">
				<Status status={req.status} />
			</div>

			{ req.reply && <>
				<h2>They replied:</h2>
				<div className="thread">
					<Status status={req.reply} />
				</div>
			</> }
			
			<div className="action-buttons">
				<MutationButton
					label={`Accept ${noun}`}
					title={`Accept ${noun}`}
					type="button"
					className="button"
					onClick={(e) => {
						e.preventDefault();
						approve(req.id);
						setLocation(backLocation);
					}}
					disabled={false}
					showError={false}
					result={approveResult}
				/>

				<MutationButton
					label={`Reject ${noun}`}
					title={`Reject ${noun}`}
					type="button"
					className="button danger"
					onClick={(e) => {
						e.preventDefault();
						reject(req.id);
						setLocation(backLocation);
					}}
					disabled={false}
					showError={false}
					result={rejectResult}
				/>
			</div>
		</>
	);
}
