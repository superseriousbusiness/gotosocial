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

package text

import (
	"context"

	"github.com/russross/blackfriday/v2"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (f *formatter) FromMarkdown(ctx context.Context, md string, mentions []*gtsmodel.Mention, tags []*gtsmodel.Tag) string {
	content := preformat(md)

	// do the markdown parsing *first*
	contentBytes := blackfriday.Run([]byte(content))

	// format tags nicely
	content = f.ReplaceTags(ctx, string(contentBytes), tags)

	// format mentions nicely
	content = f.ReplaceMentions(ctx, content, mentions)

	return postformat(content)
}
