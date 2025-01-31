// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package model

// InstanceV2 models information about this instance.
//
// swagger:model instanceV2
type InstanceV2 struct {
	// The domain of the instance.
	// example: gts.example.org
	Domain string `json:"domain"`
	// The domain of accounts on this instance.
	// This will not necessarily be the same as
	// domain.
	// example: example.org
	AccountDomain string `json:"account_domain"`
	// The title of the instance.
	// example: GoToSocial Example Instance
	Title string `json:"title"`
	// The version of GoToSocial installed on the instance.
	//
	// This will contain at least a semantic version number.
	//
	// It may also contain, after a space, the short git commit ID of the running software.
	//
	// example: 0.1.1 cb85f65
	Version string `json:"version"`
	// Whether or not instance is running in DEBUG mode. Omitted if false.
	Debug *bool `json:"debug,omitempty"`
	// The URL for the source code of the software running on this instance, in keeping with AGPL license requirements.
	// example: https://github.com/superseriousbusiness/gotosocial
	SourceURL string `json:"source_url"`
	// Description of the instance.
	//
	// Should be HTML formatted, but might be plaintext.
	//
	// This should be displayed on the 'about' page for an instance.
	Description string `json:"description"`
	// Raw (unparsed) version of description.
	DescriptionText string `json:"description_text,omitempty"`
	// Instance Custom Css
	CustomCSS string `json:"custom_css,omitempty"`
	// Basic anonymous usage data for this instance.
	Usage InstanceV2Usage `json:"usage"`
	// An image used to represent this instance.
	Thumbnail InstanceV2Thumbnail `json:"thumbnail"`
	// Primary languages of the instance + moderators/admins.
	// example: ["en"]
	Languages []string `json:"languages"`
	// Configured values and limits for this instance.
	Configuration InstanceV2Configuration `json:"configuration"`
	// Information about registering for this instance.
	Registrations InstanceV2Registrations `json:"registrations"`
	//  Hints related to contacting a representative of the instance.
	Contact InstanceV2Contact `json:"contact"`
	// An itemized list of rules for this instance.
	Rules []InstanceRule `json:"rules"`
	// Terms and conditions for accounts on this instance.
	Terms string `json:"terms,omitempty"`
	// Raw (unparsed) version of terms.
	TermsText string `json:"terms_text,omitempty"`

	// Random stats generated for the instance.
	// Only used if `instance-stats-randomize` is true.
	// Not serialized to the frontend.
	//
	// swagger:ignore
	RandomStats `json:"-"`
}

// Usage data for this instance.
//
// swagger:model instanceV2Usage
type InstanceV2Usage struct {
	Users InstanceV2Users `json:"users"`
}

// Usage data related to users on this instance.
//
// swagger:model instanceV2Users
type InstanceV2Users struct {
	// The number of active users in the past 4 weeks.
	// Currently not implemented: will always be 0.
	// example: 0
	ActiveMonth int `json:"active_month"`
}

// An image used to represent this instance.
//
// swagger:model instanceV2Thumbnail
type InstanceV2Thumbnail struct {
	// The URL for the thumbnail image.
	// example: https://example.org/fileserver/01BPSX2MKCRVMD4YN4D71G9CP5/attachment/original/01H88X0KQ2DFYYDSWYP93VDJZA.png
	URL string `json:"url"`
	// MIME type of the instance thumbnail.
	// Key/value not set if thumbnail image type unknown.
	// example: image/png
	Type string `json:"thumbnail_type,omitempty"`
	// StaticURL version of the thumbnail image.
	// example: https://example.org/fileserver/01BPSX2MKCRVMD4YN4D71G9CP5/attachment/static/01H88X0KQ2DFYYDSWYP93VDJZA.webp
	StaticURL string `json:"static_url,omitempty"`
	// MIME type of the instance thumbnail.
	// Key/value not set if thumbnail image type unknown.
	// example: image/png
	StaticType string `json:"thumbnail_static_type,omitempty"`
	// Description of the instance thumbnail.
	// Key/value not set if no description available.
	// example: picture of a cute lil' friendly sloth
	Description string `json:"thumbnail_description,omitempty"`
	// A hash computed by the BlurHash algorithm, for generating colorful preview thumbnails when media has not been downloaded yet.
	// Key/value not set if no blurhash available.
	// example: UeKUpFxuo~R%0nW;WCnhF6RjaJt757oJodS$
	Blurhash string `json:"blurhash,omitempty"`
	// Links to scaled resolution images, for high DPI screens.
	// Key/value not set if no extra versions available.
	Versions *InstanceV2ThumbnailVersions `json:"versions,omitempty"`
}

// Links to scaled resolution images, for high DPI screens.
//
// swagger:model instanceV2ThumbnailVersions
type InstanceV2ThumbnailVersions struct {
	// The URL for the thumbnail image at 1x resolution.
	// Key/value not set if scaled versions not available.
	Size1URL string `json:"@1x,omitempty"`
	// The URL for the thumbnail image at 2x resolution.
	// Key/value not set if scaled versions not available.
	Size2URL string `json:"@2x,omitempty"`
}

// InstanceV2URLs models instance-relevant URLs for client application consumption.
//
// swagger:model instanceV2URLs
type InstanceV2URLs struct {
	// Websockets address for status and notification streaming.
	// example: wss://example.org
	Streaming string `json:"streaming"`
}

// Hints related to translation.
//
// swagger:model instanceV2ConfigurationTranslation
type InstanceV2ConfigurationTranslation struct {
	// Whether the Translations API is available on this instance.
	// Not implemented so this value is always false.
	Enabled bool `json:"enabled"`
}

// Configured values and limits for this instance.
//
// swagger:model instanceV2Configuration
type InstanceV2Configuration struct {
	// URLs of interest for clients apps.
	URLs InstanceV2URLs `json:"urls"`
	// Limits related to accounts.
	Accounts InstanceConfigurationAccounts `json:"accounts"`
	// Limits related to authoring statuses.
	Statuses InstanceConfigurationStatuses `json:"statuses"`
	// Hints for which attachments will be accepted.
	MediaAttachments InstanceConfigurationMediaAttachments `json:"media_attachments"`
	// Limits related to polls.
	Polls InstanceConfigurationPolls `json:"polls"`
	// Hints related to translation.
	Translation InstanceV2ConfigurationTranslation `json:"translation"`
	// Instance configuration pertaining to emojis.
	Emojis InstanceConfigurationEmojis `json:"emojis"`
	// True if instance is running with OIDC as auth/identity backend, else omitted.
	OIDCEnabled bool `json:"oidc_enabled,omitempty"`
	// Instance VAPID configuration.
	VAPID InstanceV2ConfigurationVAPID `json:"vapid"`
}

// Information about registering for this instance.
//
// swagger:model instanceV2Registrations
type InstanceV2Registrations struct {
	// Whether registrations are enabled.
	// example: false
	Enabled bool `json:"enabled"`
	// Whether registrations require moderator approval.
	// example: true
	ApprovalRequired bool `json:"approval_required"`
	// A custom message (html string) to be shown when registrations are closed.
	// Value will be null if no message is set.
	// example: <p>Registrations are currently closed on example.org because of spam bots!</p>
	Message *string `json:"message"`
}

// Hints related to contacting a representative of the instance.
//
// swagger:model instanceV2Contact
type InstanceV2Contact struct {
	// An email address that can be messaged regarding inquiries or issues.
	// Empty string if no email address set.
	// example: someone@example.org
	Email string `json:"email"`
	// An account that can be contacted regarding inquiries or issues.
	// Key/value not present if no contact account set.
	Account *Account `json:"account,omitempty"`
}

// InstanceV2ConfigurationVAPID holds the instance's VAPID configuration.
//
// swagger:model instanceV2ConfigurationVAPID
type InstanceV2ConfigurationVAPID struct {
	// The instance's VAPID public key, Base64-encoded.
	PublicKey string `json:"public_key"`
}
