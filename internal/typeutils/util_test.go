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

package typeutils

import (
	"testing"
)

func TestMisskeyReportContentURLs1(t *testing.T) {
	content := `Note: https://bad.instance/@tobi/statuses/01GPB56GPJ37JTK9HW308HQKBQ
Note: https://bad.instance/@tobi/statuses/01GPB56GPJ37JTK9HW308HQKBQ
Note: https://bad.instance/@tobi/statuses/01GPB56GPJ37JTK9HW308HQKBQ
-----
Test report from Calckey`

	urls := misskeyReportInlineURLs(content)
	if l := len(urls); l != 3 {
		t.Fatalf("wanted 3 urls, got %d", l)
	}
}

func TestMisskeyReportContentURLs2(t *testing.T) {
	content := `This is a report
with just a normal url in it: https://example.org, and is not
misskey-formatted`

	urls := misskeyReportInlineURLs(content)
	if l := len(urls); l != 0 {
		t.Fatalf("wanted 0 urls, got %d", l)
	}
}
