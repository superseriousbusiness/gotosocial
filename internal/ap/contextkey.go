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

package ap

// ContextKey is a type used specifically for settings values on contexts within go-fed AP request chains
type ContextKey string

const (
	// ContextActivity can be used to set and retrieve the actual go-fed pub.Activity within a context.
	ContextActivity ContextKey = "activity"
	// ContextReceivingAccount can be used the set and retrieve the account being interacted with / receiving an activity in their inbox.
	ContextReceivingAccount ContextKey = "account"
	// ContextRequestingAccount can be used to set and retrieve the account of an incoming federation request.
	// This will often be the actor of the instance that's posting the request.
	ContextRequestingAccount ContextKey = "requestingAccount"
	// ContextRequestingActorIRI can be used to set and retrieve the actor of an incoming federation request.
	// This will usually be the owner of whatever activity is being posted.
	ContextRequestingActorIRI ContextKey = "requestingActorIRI"
	// ContextRequestingPublicKeyVerifier can be used to set and retrieve the public key verifier of an incoming federation request.
	ContextRequestingPublicKeyVerifier ContextKey = "requestingPublicKeyVerifier"
	// ContextRequestingPublicKeySignature can be used to set and retrieve the value of the signature header of an incoming federation request.
	ContextRequestingPublicKeySignature ContextKey = "requestingPublicKeySignature"
	// ContextFromFederatorChan can be used to pass a pointer to the fromFederator channel into the federator for use in callbacks.
	ContextFromFederatorChan ContextKey = "fromFederatorChan"
)
