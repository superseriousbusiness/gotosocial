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

package messages

import (
	"context"
	"encoding/json"
	"net/url"
	"reflect"

	"codeberg.org/gruf/go-structr"
	"codeberg.org/superseriousbusiness/activity/streams"
	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// FromClientAPI wraps a message that
// travels from the client API into the processor.
type FromClientAPI struct {

	// APObjectType ...
	APObjectType string

	// APActivityType ...
	APActivityType string

	// Optional GTS database model
	// of the Activity / Object.
	GTSModel interface{}

	// Targeted object URI.
	TargetURI string

	// Origin is the account that
	// this message originated from.
	Origin *gtsmodel.Account

	// Target is the account that
	// this message is targeting.
	Target *gtsmodel.Account
}

// fromClientAPI is an internal type
// for FromClientAPI that provides a
// json serialize / deserialize -able
// shape that minimizes required data.
type fromClientAPI struct {
	APObjectType   string          `json:"ap_object_type,omitempty"`
	APActivityType string          `json:"ap_activity_type,omitempty"`
	GTSModel       json.RawMessage `json:"gts_model,omitempty"`
	GTSModelType   string          `json:"gts_model_type,omitempty"`
	TargetURI      string          `json:"target_uri,omitempty"`
	OriginID       string          `json:"origin_id,omitempty"`
	TargetID       string          `json:"target_id,omitempty"`
}

// Serialize will serialize the worker data as data blob for storage,
// note that this will flatten some of the data e.g. only account IDs.
func (msg *FromClientAPI) Serialize() ([]byte, error) {
	var (
		modelType string
		originID  string
		targetID  string
	)

	// Set database model type if any provided.
	if t := reflect.TypeOf(msg.GTSModel); t != nil {
		modelType = t.String()
	}

	// Set origin account ID.
	if msg.Origin != nil {
		originID = msg.Origin.ID
	}

	// Set target account ID.
	if msg.Target != nil {
		targetID = msg.Target.ID
	}

	// Marshal GTS model as raw JSON block.
	modelJSON, err := json.Marshal(msg.GTSModel)
	if err != nil {
		return nil, err
	}

	// Marshal as internal JSON type.
	return json.Marshal(fromClientAPI{
		APObjectType:   msg.APObjectType,
		APActivityType: msg.APActivityType,
		GTSModel:       modelJSON,
		GTSModelType:   modelType,
		TargetURI:      msg.TargetURI,
		OriginID:       originID,
		TargetID:       targetID,
	})
}

// Deserialize will attempt to deserialize a blob of task data,
// which will involve unflattening previously serialized data and
// leave some message structures as placeholders to holding IDs.
func (msg *FromClientAPI) Deserialize(data []byte) error {
	var imsg fromClientAPI

	// Unmarshal as internal JSON type.
	err := json.Unmarshal(data, &imsg)
	if err != nil {
		return err
	}

	// Copy over the simplest fields.
	msg.APObjectType = imsg.APObjectType
	msg.APActivityType = imsg.APActivityType
	msg.TargetURI = imsg.TargetURI

	// Resolve Go type from JSON data.
	msg.GTSModel, err = resolveGTSModel(
		imsg.GTSModelType,
		imsg.GTSModel,
	)
	if err != nil {
		return err
	}

	if imsg.OriginID != "" {
		// Set origin account ID using a
		// barebones model (later filled in).
		msg.Origin = new(gtsmodel.Account)
		msg.Origin.ID = imsg.OriginID
	}

	if imsg.TargetID != "" {
		// Set target account ID using a
		// barebones model (later filled in).
		msg.Target = new(gtsmodel.Account)
		msg.Target.ID = imsg.TargetID
	}

	return nil
}

// ClientMsgIndices defines queue indices this
// message type should be accessible / stored under.
func ClientMsgIndices() []structr.IndexConfig {
	return []structr.IndexConfig{
		{Fields: "TargetURI", Multiple: true},
		{Fields: "Origin.ID", Multiple: true},
		{Fields: "Target.ID", Multiple: true},
	}
}

// FromFediAPI wraps a message that
// travels from the federating API into the processor.
type FromFediAPI struct {

	// APObjectType ...
	APObjectType string

	// APActivityType ...
	APActivityType string

	// Optional ActivityPub ID (IRI)
	// and / or model of Activity / Object.
	APIRI    *url.URL
	APObject interface{}

	// Optional GTS database model
	// of the Activity / Object.
	GTSModel interface{}

	// Targeted object URI.
	TargetURI string

	// Remote account that posted
	// this Activity to the inbox.
	Requesting *gtsmodel.Account

	// Local account which owns the inbox
	// that this Activity was posted to.
	Receiving *gtsmodel.Account
}

// fromFediAPI is an internal type
// for FromFediAPI that provides a
// json serialize / deserialize -able
// shape that minimizes required data.
type fromFediAPI struct {
	APObjectType   string                 `json:"ap_object_type,omitempty"`
	APActivityType string                 `json:"ap_activity_type,omitempty"`
	APIRI          string                 `json:"ap_iri,omitempty"`
	APObject       map[string]interface{} `json:"ap_object,omitempty"`
	GTSModel       json.RawMessage        `json:"gts_model,omitempty"`
	GTSModelType   string                 `json:"gts_model_type,omitempty"`
	TargetURI      string                 `json:"target_uri,omitempty"`
	RequestingID   string                 `json:"requesting_id,omitempty"`
	ReceivingID    string                 `json:"receiving_id,omitempty"`
}

// Serialize will serialize the worker data as data blob for storage,
// note that this will flatten some of the data e.g. only account IDs.
func (msg *FromFediAPI) Serialize() ([]byte, error) {
	var (
		gtsModelType string
		apIRI        string
		apObject     map[string]interface{}
		requestingID string
		receivingID  string
	)

	// Set AP IRI string.
	if msg.APIRI != nil {
		apIRI = msg.APIRI.String()
	}

	// Set serialized AP object data if set.
	if t, ok := msg.APObject.(vocab.Type); ok {
		obj, err := streams.Serialize(t)
		if err != nil {
			return nil, err
		}
		apObject = obj
	}

	// Set database model type if any provided.
	if t := reflect.TypeOf(msg.GTSModel); t != nil {
		gtsModelType = t.String()
	}

	// Set requesting account ID.
	if msg.Requesting != nil {
		requestingID = msg.Requesting.ID
	}

	// Set receiving account ID.
	if msg.Receiving != nil {
		receivingID = msg.Receiving.ID
	}

	// Marshal GTS model as raw JSON block.
	modelJSON, err := json.Marshal(msg.GTSModel)
	if err != nil {
		return nil, err
	}

	// Marshal as internal JSON type.
	return json.Marshal(fromFediAPI{
		APObjectType:   msg.APObjectType,
		APActivityType: msg.APActivityType,
		APIRI:          apIRI,
		APObject:       apObject,
		GTSModel:       modelJSON,
		GTSModelType:   gtsModelType,
		TargetURI:      msg.TargetURI,
		RequestingID:   requestingID,
		ReceivingID:    receivingID,
	})
}

// Deserialize will attempt to deserialize a blob of task data,
// which will involve unflattening previously serialized data and
// leave some message structures as placeholders to holding IDs.
func (msg *FromFediAPI) Deserialize(data []byte) error {
	var imsg fromFediAPI

	// Unmarshal as internal JSON type.
	err := json.Unmarshal(data, &imsg)
	if err != nil {
		return err
	}

	// Copy over the simplest fields.
	msg.APObjectType = imsg.APObjectType
	msg.APActivityType = imsg.APActivityType
	msg.TargetURI = imsg.TargetURI

	// Resolve AP object from JSON data.
	msg.APObject, err = resolveAPObject(
		imsg.APObject,
	)
	if err != nil {
		return err
	}

	// Resolve Go type from JSON data.
	msg.GTSModel, err = resolveGTSModel(
		imsg.GTSModelType,
		imsg.GTSModel,
	)
	if err != nil {
		return err
	}

	if imsg.RequestingID != "" {
		// Set requesting account ID using a
		// barebones model (later filled in).
		msg.Requesting = new(gtsmodel.Account)
		msg.Requesting.ID = imsg.RequestingID
	}

	if imsg.ReceivingID != "" {
		// Set target account ID using a
		// barebones model (later filled in).
		msg.Receiving = new(gtsmodel.Account)
		msg.Receiving.ID = imsg.ReceivingID
	}

	return nil
}

// FederatorMsgIndices defines queue indices this
// message type should be accessible / stored under.
func FederatorMsgIndices() []structr.IndexConfig {
	return []structr.IndexConfig{
		{Fields: "APIRI", Multiple: true},
		{Fields: "TargetURI", Multiple: true},
		{Fields: "Requesting.ID", Multiple: true},
		{Fields: "Receiving.ID", Multiple: true},
	}
}

// resolveAPObject resolves an ActivityPub object from its "serialized" JSON map
// (yes the terminology here is weird, but that's how go-fed/activity is written).
func resolveAPObject(data map[string]interface{}) (interface{}, error) {
	if len(data) == 0 {
		// No data given.
		return nil, nil
	}

	// Resolve vocab.Type from "raw" input data map.
	return streams.ToType(context.Background(), data)
}

// resolveGTSModel is unfortunately where things get messy... our data is stored as JSON
// in the database, which serializes struct types as key-value pairs surrounded by curly
// braces. Deserializing from that gives us back a data blob of key-value pairs, which
// we then need to wrangle back into the original type. So we also store the type name
// and use this to determine the appropriate Go structure type to unmarshal into to.
func resolveGTSModel(typ string, data []byte) (interface{}, error) {
	if typ == "" {
		// No data given.
		return nil, nil
	}

	var value interface{}

	switch typ {
	case reflect.TypeOf((*gtsmodel.Account)(nil)).String():
		value = new(gtsmodel.Account)
	case reflect.TypeOf((*gtsmodel.Block)(nil)).String():
		value = new(gtsmodel.Block)
	case reflect.TypeOf((*gtsmodel.Follow)(nil)).String():
		value = new(gtsmodel.Follow)
	case reflect.TypeOf((*gtsmodel.FollowRequest)(nil)).String():
		value = new(gtsmodel.FollowRequest)
	case reflect.TypeOf((*gtsmodel.Move)(nil)).String():
		value = new(gtsmodel.Move)
	case reflect.TypeOf((*gtsmodel.Poll)(nil)).String():
		value = new(gtsmodel.Poll)
	case reflect.TypeOf((*gtsmodel.PollVote)(nil)).String():
		value = new(*gtsmodel.PollVote)
	case reflect.TypeOf((*gtsmodel.Report)(nil)).String():
		value = new(gtsmodel.Report)
	case reflect.TypeOf((*gtsmodel.Status)(nil)).String():
		value = new(gtsmodel.Status)
	case reflect.TypeOf((*gtsmodel.StatusFave)(nil)).String():
		value = new(gtsmodel.StatusFave)
	default:
		return nil, gtserror.Newf("unknown type: %s", typ)
	}

	// Attempt to unmarshal value JSON into destination.
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, gtserror.Newf("error unmarshaling %s value data: %w", typ, err)
	}

	return value, nil
}
