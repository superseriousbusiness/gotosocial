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

package text

import (
	"context"
	"fmt"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *formatter) FromPlain(ctx context.Context, plain string, mentions []*gtsmodel.Mention, tags []*gtsmodel.Tag) string {
	content := preformat(plain)

	// sanitize any html elements
	content = RemoveHTML(content)

	// format links nicely
	content = f.ReplaceLinks(ctx, content)

	// format tags nicely
	content = f.ReplaceTags(ctx, content, tags)

	// format mentions nicely
	content = f.ReplaceMentions(ctx, content, mentions)

	// replace newlines with breaks
	content = strings.ReplaceAll(content, "\n", "<br />")

	// wrap the whole thing in a pee
	content = fmt.Sprintf(`<p>%s</p>`, content)

	return postformat(content)
}
