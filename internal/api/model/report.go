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

package model

// Report models a moderation report submitted to the instance, either via the client API or via the federated API.
//
// swagger:model report
type Report struct {
	// ID of the report.
	// example: 01FBVD42CQ3ZEEVMW180SBX03B
	ID string `json:"id"`
	// The date when this report was created (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	CreatedAt string `json:"created_at"`
	// Whether an action has been taken by an admin in response to this report.
	// example: false
	ActionTaken bool `json:"action_taken"`
	// If an action was taken, at what time was this done? (ISO 8601 Datetime)
	// Will be null if not set / no action yet taken.
	// example: 2021-07-30T09:20:25+00:00
	ActionTakenAt *string `json:"action_taken_at"`
	// If an action was taken, what comment was made by the admin on the taken action?
	// Will be null if not set / no action yet taken.
	// example: Account was suspended.
	ActionTakenComment *string `json:"action_taken_comment"`
	// Under what category was this report created?
	// example: spam
	Category string `json:"category"`
	// Comment submitted when the report was created.
	// Will be empty if no comment was submitted.
	// example: This person has been harassing me.
	Comment string `json:"comment"`
	// Bool to indicate that report should be federated to remote instance.
	// example: true
	Forwarded bool `json:"forwarded"`
	// Array of IDs of statuses that were submitted along with this report.
	// Will be empty if no status IDs were submitted.
	// example: ["01GPBN5YDY6JKBWE44H7YQBDCQ","01GPBN65PDWSBPWVDD0SQCFFY3"]
	StatusIDs []string `json:"status_ids"`
	// Array of rule IDs that were submitted along with this report.
	// Will be empty if no rule IDs were submitted.
	// example: [1, 2]
	RuleIDs []int `json:"rule_ids"`
	// Account that was reported.
	TargetAccount *Account `json:"target_account"`
}

// ReportCreateRequest models user report creation parameters.
//
// swagger:parameters reportCreate
type ReportCreateRequest struct {
	// ID of the account to report.
	// example: 01GPE75FXSH2EGFBF85NXPH3KP
	// in: formData
	// required: true
	AccountID string `form:"account_id" json:"account_id" xml:"account_id"`
	// IDs of statuses to attach to the report to provide additional context.
	// example: ["01GPE76N4SBVRZ8K24TW51ZZQ4","01GPE76WN9JZE62EPT3Q9FRRD4"]
	// in: formData
	StatusIDs []string `form:"status_ids[]" json:"status_ids" xml:"status_ids"`
	// The reason for the report. Default maximum of 1000 characters.
	// example: Anti-Blackness, transphobia.
	// in: formData
	Comment string `form:"comment" json:"comment" xml:"comment"`
	// If the account is remote, should the report be forwarded to the remote admin?
	// example: true
	// default: false
	// in: formData
	Forward bool `form:"forward" json:"forward" xml:"forward"`
	// Specify if the report is due to spam, violation of enumerated instance rules, or some other reason.
	// Currently only 'other' is supported.
	// example: other
	// default: other
	// in: formData
	Category string `form:"category" json:"category" xml:"category"`
	// IDs of rules on this instance which have been broken according to the reporter.
	// This is currently not supported, provided only for API compatibility.
	// example: [1, 2, 3]
	// in: formData
	RuleIDs []int `form:"rule_ids[]" json:"rule_ids" xml:"rule_ids"`
}
