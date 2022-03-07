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

package flag

import (
	"github.com/spf13/cobra"
	"github.com/superseriousbusiness/gotosocial/internal/config"
)

// Server attaches all flags pertaining to running the GtS server or testrig.
func Server(cmd *cobra.Command, values config.Values) {
	Template(cmd, values)
	Accounts(cmd, values)
	Media(cmd, values)
	Storage(cmd, values)
	Statuses(cmd, values)
	LetsEncrypt(cmd, values)
	OIDC(cmd, values)
	SMTP(cmd, values)
	Router(cmd, values)
	Syslog(cmd, values)
}

// Router attaches flags pertaining to the gin router.
func Router(cmd *cobra.Command, values config.Values) {
	cmd.PersistentFlags().String(config.Keys.BindAddress, values.BindAddress, usage.BindAddress)
	cmd.PersistentFlags().Int(config.Keys.Port, values.Port, usage.Port)
	cmd.PersistentFlags().StringSlice(config.Keys.TrustedProxies, values.TrustedProxies, usage.TrustedProxies)
}

// Template attaches flags pertaining to templating config.
func Template(cmd *cobra.Command, values config.Values) {
	cmd.Flags().String(config.Keys.WebTemplateBaseDir, values.WebTemplateBaseDir, usage.WebTemplateBaseDir)
	cmd.Flags().String(config.Keys.WebAssetBaseDir, values.WebAssetBaseDir, usage.WebAssetBaseDir)
}

// Accounts attaches flags pertaining to account config.
func Accounts(cmd *cobra.Command, values config.Values) {
	cmd.Flags().Bool(config.Keys.AccountsRegistrationOpen, values.AccountsRegistrationOpen, usage.AccountsRegistrationOpen)
	cmd.Flags().Bool(config.Keys.AccountsApprovalRequired, values.AccountsApprovalRequired, usage.AccountsApprovalRequired)
	cmd.Flags().Bool(config.Keys.AccountsReasonRequired, values.AccountsReasonRequired, usage.AccountsReasonRequired)
}

// Media attaches flags pertaining to media config.
func Media(cmd *cobra.Command, values config.Values) {
	cmd.Flags().Int(config.Keys.MediaImageMaxSize, values.MediaImageMaxSize, usage.MediaImageMaxSize)
	cmd.Flags().Int(config.Keys.MediaVideoMaxSize, values.MediaVideoMaxSize, usage.MediaVideoMaxSize)
	cmd.Flags().Int(config.Keys.MediaDescriptionMinChars, values.MediaDescriptionMinChars, usage.MediaDescriptionMinChars)
	cmd.Flags().Int(config.Keys.MediaDescriptionMaxChars, values.MediaDescriptionMaxChars, usage.MediaDescriptionMaxChars)
	cmd.Flags().Int(config.Keys.MediaRemoteCacheDays, values.MediaRemoteCacheDays, usage.MediaRemoteCacheDays)
}

// Storage attaches flags pertaining to storage config.
func Storage(cmd *cobra.Command, values config.Values) {
	cmd.Flags().String(config.Keys.StorageBackend, values.StorageBackend, usage.StorageBackend)
	cmd.Flags().String(config.Keys.StorageLocalBasePath, values.StorageLocalBasePath, usage.StorageLocalBasePath)
}

// Statuses attaches flags pertaining to statuses config.
func Statuses(cmd *cobra.Command, values config.Values) {
	cmd.Flags().Int(config.Keys.StatusesMaxChars, values.StatusesMaxChars, usage.StatusesMaxChars)
	cmd.Flags().Int(config.Keys.StatusesCWMaxChars, values.StatusesCWMaxChars, usage.StatusesCWMaxChars)
	cmd.Flags().Int(config.Keys.StatusesPollMaxOptions, values.StatusesPollMaxOptions, usage.StatusesPollMaxOptions)
	cmd.Flags().Int(config.Keys.StatusesPollOptionMaxChars, values.StatusesPollOptionMaxChars, usage.StatusesPollOptionMaxChars)
	cmd.Flags().Int(config.Keys.StatusesMediaMaxFiles, values.StatusesMediaMaxFiles, usage.StatusesMediaMaxFiles)
}

// LetsEncrypt attaches flags pertaining to letsencrypt config.
func LetsEncrypt(cmd *cobra.Command, values config.Values) {
	cmd.Flags().Bool(config.Keys.LetsEncryptEnabled, values.LetsEncryptEnabled, usage.LetsEncryptEnabled)
	cmd.Flags().Int(config.Keys.LetsEncryptPort, values.LetsEncryptPort, usage.LetsEncryptPort)
	cmd.Flags().String(config.Keys.LetsEncryptCertDir, values.LetsEncryptCertDir, usage.LetsEncryptCertDir)
	cmd.Flags().String(config.Keys.LetsEncryptEmailAddress, values.LetsEncryptEmailAddress, usage.LetsEncryptEmailAddress)
}

// OIDC attaches flags pertaining to oidc config.
func OIDC(cmd *cobra.Command, values config.Values) {
	cmd.Flags().Bool(config.Keys.OIDCEnabled, values.OIDCEnabled, usage.OIDCEnabled)
	cmd.Flags().String(config.Keys.OIDCIdpName, values.OIDCIdpName, usage.OIDCIdpName)
	cmd.Flags().Bool(config.Keys.OIDCSkipVerification, values.OIDCSkipVerification, usage.OIDCSkipVerification)
	cmd.Flags().String(config.Keys.OIDCIssuer, values.OIDCIssuer, usage.OIDCIssuer)
	cmd.Flags().String(config.Keys.OIDCClientID, values.OIDCClientID, usage.OIDCClientID)
	cmd.Flags().String(config.Keys.OIDCClientSecret, values.OIDCClientSecret, usage.OIDCClientSecret)
	cmd.Flags().StringSlice(config.Keys.OIDCScopes, values.OIDCScopes, usage.OIDCScopes)
}

// SMTP attaches flags pertaining to smtp/email config.
func SMTP(cmd *cobra.Command, values config.Values) {
	cmd.Flags().String(config.Keys.SMTPHost, values.SMTPHost, usage.SMTPHost)
	cmd.Flags().Int(config.Keys.SMTPPort, values.SMTPPort, usage.SMTPPort)
	cmd.Flags().String(config.Keys.SMTPUsername, values.SMTPUsername, usage.SMTPUsername)
	cmd.Flags().String(config.Keys.SMTPPassword, values.SMTPPassword, usage.SMTPPassword)
	cmd.Flags().String(config.Keys.SMTPFrom, values.SMTPFrom, usage.SMTPFrom)
}

// Syslog attaches flags pertaining to syslog config.
func Syslog(cmd *cobra.Command, values config.Values) {
	cmd.Flags().Bool(config.Keys.SyslogEnabled, values.SyslogEnabled, usage.SyslogEnabled)
	cmd.Flags().String(config.Keys.SyslogProtocol, values.SyslogProtocol, usage.SyslogProtocol)
	cmd.Flags().String(config.Keys.SyslogAddress, values.SyslogAddress, usage.SyslogAddress)
}
