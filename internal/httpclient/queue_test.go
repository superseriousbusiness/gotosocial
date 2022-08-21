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

package httpclient

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type QueueTestSuite struct {
	suite.Suite
}

func (suite *QueueTestSuite) TestQueue() {
	maxOpenConns := 5
	waitTimeout := 1 * time.Second

	rc := &requestQueue{
		maxOpenConns: maxOpenConns,
	}

	// fill all the open connections
	var done func()
	for i, n := range make([]interface{}, maxOpenConns) {
		w, d := rc.getWaitSpot("example.org", http.MethodPost)
		w <- n
		if i == maxOpenConns-1 {
			// save the last done function
			done = d
		}
	}

	// try to wait again for the same host/method combo, it should timeout
	waitAgain, _ := rc.getWaitSpot("example.org", "post")

	select {
	case waitAgain <- struct{}{}:
		suite.FailNow("first wait did not time out")
	case <-time.After(waitTimeout):
		break
	}

	// now close the final done that we derived earlier
	done()

	// try waiting again, it should work this time
	select {
	case waitAgain <- struct{}{}:
		break
	case <-time.After(waitTimeout):
		suite.FailNow("second wait timed out")
	}

	// the POST queue is now sitting on full
	suite.Len(waitAgain, maxOpenConns)

	// we should still be able to make a GET for the same host though
	getWait, getDone := rc.getWaitSpot("example.org", http.MethodGet)
	select {
	case getWait <- struct{}{}:
		break
	case <-time.After(waitTimeout):
		suite.FailNow("get wait timed out")
	}

	// the GET queue has one request waiting
	suite.Len(getWait, 1)
	// clear it...
	getDone()
	suite.Empty(getWait)

	// even though the POST queue for example.org is full, we
	// should still be able to make a POST request to another host :)
	waitForAnotherHost, _ := rc.getWaitSpot("somewhere.else", http.MethodPost)
	select {
	case waitForAnotherHost <- struct{}{}:
		break
	case <-time.After(waitTimeout):
		suite.FailNow("get wait timed out")
	}

	suite.Len(waitForAnotherHost, 1)
}

func TestQueueTestSuite(t *testing.T) {
	suite.Run(t, &QueueTestSuite{})
}
