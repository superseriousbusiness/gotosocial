#!/bin/sh

# SPDX-FileCopyrightText: 2023 GoToSocial Authors <admin@gotosocial.org>
#
# SPDX-License-Identifier: AGPL-3.0-only

# this script is really just here because GoReleaser doesn't let
# you set env vars in your 'before' commands in the free version

set -eu

BUDO_BUILD=1 node web/source
