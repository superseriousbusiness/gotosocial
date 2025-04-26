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

package admin

import (
	"context"
	"fmt"
	"slices"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
	"code.superseriousbusiness.org/gotosocial/internal/transport"
	"code.superseriousbusiness.org/gotosocial/internal/transport/delivery"
)

// NOTE:
// Having these functions in the processor, which is
// usually the intermediary that performs *processing*
// between the HTTP route handlers and the underlying
// database / storage layers is a little odd, so this
// may be subject to change!
//
// For now at least, this is a useful place that has
// access to the underlying database, workers and
// causes no dependency cycles with this use case!

// FillWorkerQueues recovers all serialized worker tasks from the database
// (if any!), and pushes them to each of their relevant worker queues.
func (p *Processor) FillWorkerQueues(ctx context.Context) error {
	log.Info(ctx, "rehydrate!")

	// Get all persisted worker tasks from db.
	//
	// (database returns these as ASCENDING, i.e.
	// returned in the order they were inserted).
	tasks, err := p.state.DB.GetWorkerTasks(ctx)
	if err != nil {
		return gtserror.Newf("error fetching worker tasks from db: %w", err)
	}

	var (
		// Counts of each task type
		// successfully recovered.
		delivery  int
		federator int
		client    int

		// Failed recoveries.
		errors int
	)

loop:

	// Handle each persisted task, removing
	// all those we can't handle. Leaving us
	// with a slice of tasks we can safely
	// delete from being persisted in the DB.
	for i := 0; i < len(tasks); {
		var err error

		// Task at index.
		task := tasks[i]

		// Appropriate task count
		// pointer to increment.
		var counter *int

		// Attempt to recovery persisted
		// task depending on worker type.
		switch task.WorkerType {
		case gtsmodel.DeliveryWorker:
			err = p.pushDelivery(ctx, task)
			counter = &delivery
		case gtsmodel.FederatorWorker:
			err = p.pushFederator(ctx, task)
			counter = &federator
		case gtsmodel.ClientWorker:
			err = p.pushClient(ctx, task)
			counter = &client
		default:
			err = fmt.Errorf("invalid worker type %d", task.WorkerType)
		}

		if err != nil {
			log.Errorf(ctx, "error pushing task %d: %v", task.ID, err)

			// Drop error'd task from slice.
			tasks = slices.Delete(tasks, i, i+1)

			// Incr errors.
			errors++
			continue loop
		}

		// Increment slice
		// index & counter.
		(*counter)++
		i++
	}

	// Tasks that worker successfully pushed
	// to their appropriate workers, we can
	// safely now remove from the database.
	for _, task := range tasks {
		if err := p.state.DB.DeleteWorkerTaskByID(ctx, task.ID); err != nil {
			log.Errorf(ctx, "error deleting task from db: %v", err)
		}
	}

	// Log recovered tasks.
	log.WithContext(ctx).
		WithField("delivery", delivery).
		WithField("federator", federator).
		WithField("client", client).
		WithField("errors", errors).
		Info("recovered queued tasks")

	return nil
}

// PersistWorkerQueues pops all queued worker tasks (that are themselves persistable, i.e. not
// dereference tasks which are just function ptrs), serializes and persists them to the database.
func (p *Processor) PersistWorkerQueues(ctx context.Context) error {
	log.Info(ctx, "dehydrate!")

	var (
		// Counts of each task type
		// successfully persisted.
		delivery  int
		federator int
		client    int

		// Failed persists.
		errors int

		// Serialized tasks to persist.
		tasks []*gtsmodel.WorkerTask
	)

	for {
		// Pop all queued deliveries.
		task, err := p.popDelivery()
		if err != nil {
			log.Errorf(ctx, "error popping delivery: %v", err)
			errors++ // incr error count.
			continue
		}

		if task == nil {
			// No more queue
			// tasks to pop!
			break
		}

		// Append serialized task.
		tasks = append(tasks, task)
		delivery++ // incr count
	}

	for {
		// Pop queued federator msgs.
		task, err := p.popFederator()
		if err != nil {
			log.Errorf(ctx, "error popping federator message: %v", err)
			errors++ // incr count
			continue
		}

		if task == nil {
			// No more queue
			// tasks to pop!
			break
		}

		// Append serialized task.
		tasks = append(tasks, task)
		federator++ // incr count
	}

	for {
		// Pop queued client msgs.
		task, err := p.popClient()
		if err != nil {
			log.Errorf(ctx, "error popping client message: %v", err)
			continue
		}

		if task == nil {
			// No more queue
			// tasks to pop!
			break
		}

		// Append serialized task.
		tasks = append(tasks, task)
		client++ // incr count
	}

	// Persist all serialized queued worker tasks to database.
	if err := p.state.DB.PutWorkerTasks(ctx, tasks); err != nil {
		return gtserror.Newf("error putting tasks in db: %w", err)
	}

	// Log recovered tasks.
	log.WithContext(ctx).
		WithField("delivery", delivery).
		WithField("federator", federator).
		WithField("client", client).
		WithField("errors", errors).
		Info("persisted queued tasks")

	return nil
}

// pushDelivery parses a valid delivery.Delivery{} from serialized task data and pushes to queue.
func (p *Processor) pushDelivery(ctx context.Context, task *gtsmodel.WorkerTask) error {
	dlv := new(delivery.Delivery)

	// Deserialize the raw worker task data into delivery.
	if err := dlv.Deserialize(task.TaskData); err != nil {
		return gtserror.Newf("error deserializing delivery: %w", err)
	}

	var tsport transport.Transport

	if uri := dlv.ActorID; uri != "" {
		// Fetch the actor account by provided URI from db.
		account, err := p.state.DB.GetAccountByURI(ctx, uri)
		if err != nil {
			return gtserror.Newf("error getting actor account %s from db: %w", uri, err)
		}

		// Fetch a transport for request signing for actor's account username.
		tsport, err = p.transport.NewTransportForUsername(ctx, account.Username)
		if err != nil {
			return gtserror.Newf("error getting transport for actor %s: %w", uri, err)
		}
	} else {
		var err error

		// No actor was given, will be signed by instance account.
		tsport, err = p.transport.NewTransportForUsername(ctx, "")
		if err != nil {
			return gtserror.Newf("error getting instance account transport: %w", err)
		}
	}

	// Using transport, add actor signature to delivery.
	if err := tsport.SignDelivery(dlv); err != nil {
		return gtserror.Newf("error signing delivery: %w", err)
	}

	// Push deserialized task to delivery queue.
	p.state.Workers.Delivery.Queue.Push(dlv)

	return nil
}

// popDelivery pops delivery.Delivery{} from queue and serializes as valid task data.
func (p *Processor) popDelivery() (*gtsmodel.WorkerTask, error) {

	// Pop waiting delivery from the delivery worker.
	delivery, ok := p.state.Workers.Delivery.Queue.Pop()
	if !ok {
		return nil, nil
	}

	// Serialize the delivery task data.
	data, err := delivery.Serialize()
	if err != nil {
		return nil, gtserror.Newf("error serializing delivery: %w", err)
	}

	return &gtsmodel.WorkerTask{
		// ID is autoincrement
		WorkerType: gtsmodel.DeliveryWorker,
		TaskData:   data,
		CreatedAt:  time.Now(),
	}, nil
}

// pushClient parses a valid messages.FromFediAPI{} from serialized task data and pushes to queue.
func (p *Processor) pushFederator(ctx context.Context, task *gtsmodel.WorkerTask) error {
	var msg messages.FromFediAPI

	// Deserialize the raw worker task data into message.
	if err := msg.Deserialize(task.TaskData); err != nil {
		return gtserror.Newf("error deserializing federator message: %w", err)
	}

	if rcv := msg.Receiving; rcv != nil {
		// Only a placeholder receiving account will be populated,
		// fetch the actual model from database by persisted ID.
		account, err := p.state.DB.GetAccountByID(ctx, rcv.ID)
		if err != nil {
			return gtserror.Newf("error fetching receiving account %s from db: %w", rcv.ID, err)
		}

		// Set the now populated
		// receiving account model.
		msg.Receiving = account
	}

	if req := msg.Requesting; req != nil {
		// Only a placeholder requesting account will be populated,
		// fetch the actual model from database by persisted ID.
		account, err := p.state.DB.GetAccountByID(ctx, req.ID)
		if err != nil {
			return gtserror.Newf("error fetching requesting account %s from db: %w", req.ID, err)
		}

		// Set the now populated
		// requesting account model.
		msg.Requesting = account
	}

	// Push populated task to the federator queue.
	p.state.Workers.Federator.Queue.Push(&msg)

	return nil
}

// popFederator pops messages.FromFediAPI{} from queue and serializes as valid task data.
func (p *Processor) popFederator() (*gtsmodel.WorkerTask, error) {

	// Pop waiting message from the federator worker.
	msg, ok := p.state.Workers.Federator.Queue.Pop()
	if !ok {
		return nil, nil
	}

	// Serialize message task data.
	data, err := msg.Serialize()
	if err != nil {
		return nil, gtserror.Newf("error serializing federator message: %w", err)
	}

	return &gtsmodel.WorkerTask{
		// ID is autoincrement
		WorkerType: gtsmodel.FederatorWorker,
		TaskData:   data,
		CreatedAt:  time.Now(),
	}, nil
}

// pushClient parses a valid messages.FromClientAPI{} from serialized task data and pushes to queue.
func (p *Processor) pushClient(ctx context.Context, task *gtsmodel.WorkerTask) error {
	var msg messages.FromClientAPI

	// Deserialize the raw worker task data into message.
	if err := msg.Deserialize(task.TaskData); err != nil {
		return gtserror.Newf("error deserializing client message: %w", err)
	}

	if org := msg.Origin; org != nil {
		// Only a placeholder origin account will be populated,
		// fetch the actual model from database by persisted ID.
		account, err := p.state.DB.GetAccountByID(ctx, org.ID)
		if err != nil {
			return gtserror.Newf("error fetching origin account %s from db: %w", org.ID, err)
		}

		// Set the now populated
		// origin account model.
		msg.Origin = account
	}

	if trg := msg.Target; trg != nil {
		// Only a placeholder target account will be populated,
		// fetch the actual model from database by persisted ID.
		account, err := p.state.DB.GetAccountByID(ctx, trg.ID)
		if err != nil {
			return gtserror.Newf("error fetching target account %s from db: %w", trg.ID, err)
		}

		// Set the now populated
		// target account model.
		msg.Target = account
	}

	// Push populated task to the federator queue.
	p.state.Workers.Client.Queue.Push(&msg)

	return nil
}

// popClient pops messages.FromClientAPI{} from queue and serializes as valid task data.
func (p *Processor) popClient() (*gtsmodel.WorkerTask, error) {

	// Pop waiting message from the client worker.
	msg, ok := p.state.Workers.Client.Queue.Pop()
	if !ok {
		return nil, nil
	}

	// Serialize message task data.
	data, err := msg.Serialize()
	if err != nil {
		return nil, gtserror.Newf("error serializing client message: %w", err)
	}

	return &gtsmodel.WorkerTask{
		// ID is autoincrement
		WorkerType: gtsmodel.ClientWorker,
		TaskData:   data,
		CreatedAt:  time.Now(),
	}, nil
}
