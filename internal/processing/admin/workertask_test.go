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

package admin_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/httpclient"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/transport/delivery"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

var (
	testDeliveries = []*delivery.Delivery{
		{
			PubKeyID: "https://google.com/users/bigboy#pubkey",
			ActorID:  "https://google.com/users/bigboy",
			ObjectID: "https://google.com/users/bigboy/follow/1",
			TargetID: "https://askjeeves.com/users/smallboy",
			Request:  toRequest("POST", "https://askjeeves.com/users/smallboy/inbox", []byte("data!")),
		},
		{
			Request: toRequest("GET", "https://google.com", []byte("uwu im just a wittle seawch engwin")),
		},
	}

	testFederatorMsgs = []*messages.FromFediAPI{
		{
			APObjectType:   ap.ObjectNote,
			APActivityType: ap.ActivityCreate,
			TargetURI:      "https://gotosocial.org",
			Requesting:     &gtsmodel.Account{ID: "654321"},
			Receiving:      &gtsmodel.Account{ID: "123456"},
		},
		{
			APObjectType:   ap.ObjectProfile,
			APActivityType: ap.ActivityUpdate,
			TargetURI:      "https://uk-queen-is-dead.org",
			Requesting:     &gtsmodel.Account{ID: "123456"},
			Receiving:      &gtsmodel.Account{ID: "654321"},
		},
	}

	testClientMsgs = []*messages.FromClientAPI{
		{
			APObjectType:   ap.ObjectNote,
			APActivityType: ap.ActivityCreate,
			TargetURI:      "https://gotosocial.org",
			Origin:         &gtsmodel.Account{ID: "654321"},
			Target:         &gtsmodel.Account{ID: "123456"},
		},
		{
			APObjectType:   ap.ObjectProfile,
			APActivityType: ap.ActivityUpdate,
			TargetURI:      "https://uk-queen-is-dead.org",
			Origin:         &gtsmodel.Account{ID: "123456"},
			Target:         &gtsmodel.Account{ID: "654321"},
		},
	}
)

type WorkerTaskTestSuite struct {
	AdminStandardTestSuite
}

func (suite *WorkerTaskTestSuite) TestFillWorkerQueues() {
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Fill the worker queues from persisted task data.
	err := suite.adminProcessor.FillWorkerQueues(ctx)
	suite.NoError(err)
}

func (suite *WorkerTaskTestSuite) TestPersistWorkerQueues() {
	ctx, cncl := context.WithCancel(context.Background())
	defer cncl()

	// Push all test worker tasks to their respective queues.
	suite.state.Workers.Delivery.Queue.Push(testDeliveries...)
	suite.state.Workers.Federator.Queue.Push(testFederatorMsgs...)
	suite.state.Workers.Client.Queue.Push(testClientMsgs...)

	// Persist the worker queued tasks to database.
	err := suite.adminProcessor.PersistWorkerQueues(ctx)
	suite.NoError(err)

	// Fetch all the persisted tasks from database.
	tasks, err := suite.state.DB.GetWorkerTasks(ctx)
	suite.NoError(err)

	var (
		// Deserialized
		// task counts.
		ndelivery  int
		nfederator int
		nclient    int
	)

	// Check persisted task data.
	for _, task := range tasks {
		switch task.WorkerType {
		case gtsmodel.DeliveryWorker:
			var dlv delivery.Delivery

			// Incr count.
			ndelivery++

			// Deserialize the persisted task data.
			err := dlv.Deserialize(task.TaskData)
			suite.NoError(err)

			// Check that we have this delivery in slice.
			err = containsSerializable(testDeliveries, &dlv)
			suite.NoError(err)

		case gtsmodel.FederatorWorker:
			var msg messages.FromFediAPI

			// Incr count.
			nfederator++

			// Deserialize the persisted task data.
			err := msg.Deserialize(task.TaskData)
			suite.NoError(err)

			// Check that we have this message in slice.
			err = containsSerializable(testFederatorMsgs, &msg)
			suite.NoError(err)

		case gtsmodel.ClientWorker:
			var msg messages.FromClientAPI

			// Incr count.
			nclient++

			// Deserialize the persisted task data.
			err := msg.Deserialize(task.TaskData)
			suite.NoError(err)

			// Check that we have this message in slice.
			err = containsSerializable(testClientMsgs, &msg)
			suite.NoError(err)

		default:
			suite.T().Errorf("unexpected worker type: %d", task.WorkerType)
		}
	}

	// Ensure persisted task counts as expected.
	suite.Equal(len(testDeliveries), ndelivery)
	suite.Equal(len(testFederatorMsgs), nfederator)
	suite.Equal(len(testClientMsgs), nclient)
}

func (suite *WorkerTaskTestSuite) SetupTest() {
	suite.AdminStandardTestSuite.SetupTest()
	// we don't want workers running
	testrig.StopWorkers(&suite.state)
}

func TestWorkerTaskTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerTaskTestSuite))
}

// containsSerializeable returns whether slice of serializables contains given serializable entry.
func containsSerializable[T interface{ Serialize() ([]byte, error) }](expect []T, have T) error {
	// Serialize wanted value.
	bh, err := have.Serialize()
	if err != nil {
		panic(err)
	}

	var strings []string

	for _, t := range expect {
		// Serialize expected value.
		be, err := t.Serialize()
		if err != nil {
			panic(err)
		}

		// Alloc as string.
		se := string(be)

		if se == string(bh) {
			// We have this entry!
			return nil
		}

		// Add to serialized strings.
		strings = append(strings, se)
	}

	return fmt.Errorf("could not find %s in %s", string(bh), strings)
}

// urlStr simply returns u.String() or "" if nil.
func urlStr(u *url.URL) string {
	if u == nil {
		return ""
	}
	return u.String()
}

// accountID simply returns account.ID or "" if nil.
func accountID(account *gtsmodel.Account) string {
	if account == nil {
		return ""
	}
	return account.ID
}

// toRequest creates httpclient.Request from HTTP method, URL and body data.
func toRequest(method string, url string, body []byte) httpclient.Request {
	var rbody io.Reader
	if body != nil {
		rbody = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, rbody)
	if err != nil {
		panic(err)
	}
	return httpclient.WrapRequest(req)
}

// toJSON marshals input type as JSON data.
func toJSON(a any) []byte {
	b, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}
	return b
}
