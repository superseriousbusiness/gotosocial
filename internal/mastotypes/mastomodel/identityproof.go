/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package mastotypes

// IdentityProof represents a proof from an external identity provider. See https://docs.joinmastodon.org/entities/identityproof/
type IdentityProof struct {
	// The name of the identity provider.
	Provider string `json:"provider"`
	// The account owner's username on the identity provider's service.
	ProviderUsername string `json:"provider_username"`
	// The account owner's profile URL on the identity provider.
	ProfileURL string `json:"profile_url"`
	// A link to a statement of identity proof, hosted by the identity provider.
	ProofURL string `json:"proof_url"`
	// When the identity proof was last updated.
	UpdatedAt string `json:"updated_at"`
}
