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

package media

import (
	"fmt"
	"strconv"
	"strings"
)

func parseFocus(focus string) (focusx, focusy float32, err error) {
	if focus == "" {
		return
	}
	spl := strings.Split(focus, ",")
	if len(spl) != 2 {
		err = fmt.Errorf("improperly formatted focus %s", focus)
		return
	}
	xStr := spl[0]
	yStr := spl[1]
	if xStr == "" || yStr == "" {
		err = fmt.Errorf("improperly formatted focus %s", focus)
		return
	}
	fx, err := strconv.ParseFloat(xStr, 32)
	if err != nil {
		err = fmt.Errorf("improperly formatted focus %s: %s", focus, err)
		return
	}
	if fx > 1 || fx < -1 {
		err = fmt.Errorf("improperly formatted focus %s", focus)
		return
	}
	focusx = float32(fx)
	fy, err := strconv.ParseFloat(yStr, 32)
	if err != nil {
		err = fmt.Errorf("improperly formatted focus %s: %s", focus, err)
		return
	}
	if fy > 1 || fy < -1 {
		err = fmt.Errorf("improperly formatted focus %s", focus)
		return
	}
	focusy = float32(fy)
	return
}
