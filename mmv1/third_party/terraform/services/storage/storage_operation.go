// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-provider-google/google/tpgresource"
	transport_tpg "github.com/hashicorp/terraform-provider-google/google/transport"
)

type StorageOperationWaiter struct {
	Config    *transport_tpg.Config
	UserAgent string
	SelfLink  string
	tpgresource.CommonOperationWaiter
}

func (w *StorageOperationWaiter) QueryOp() (interface{}, error) {
	if w == nil {
		return nil, fmt.Errorf("Cannot query operation, it's unset or nil.")
	}
	// Returns the proper get.
	url := fmt.Sprintf(w.SelfLink)

	return transport_tpg.SendRequest(transport_tpg.SendRequestOptions{
		Config:    w.Config,
		Method:    "GET",
		RawURL:    url,
		UserAgent: w.UserAgent,
	})
}

func createStorageWaiter(config *transport_tpg.Config, op map[string]interface{}, activity, userAgent string) (*StorageOperationWaiter, error) {
	val, ok := op["selfLink"].(string)
	if !ok {
		return nil, fmt.Errorf("Unable to parse selfLink from LRO metadata")
	}
	w := &StorageOperationWaiter{
		Config:    config,
		UserAgent: userAgent,
		SelfLink:  val,
	}
	if err := w.CommonOperationWaiter.SetOp(op); err != nil {
		return nil, err
	}
	return w, nil
}

// nolint: deadcode,unused
func StorageOperationWaitTimeWithResponse(config *transport_tpg.Config, op map[string]interface{}, response *map[string]interface{}, activity, userAgent string, timeout time.Duration) error {
	w, err := createStorageWaiter(config, op, activity, userAgent)
	if err != nil {
		return err
	}
	if err := tpgresource.OperationWait(w, activity, timeout, config.PollInterval); err != nil {
		return err
	}
	rawResponse := []byte(w.CommonOperationWaiter.Op.Response)
	if len(rawResponse) == 0 {
		return errors.New("`resource` not set in operation response")
	}
	return json.Unmarshal(rawResponse, response)
}

func StorageOperationWaitTime(config *transport_tpg.Config, op map[string]interface{}, activity, userAgent string, timeout time.Duration) error {
	if val, ok := op["name"]; !ok || val == "" {
		// This was a synchronous call - there is no operation to wait for.
		return nil
	}
	w, err := createStorageWaiter(config, op, activity, userAgent)
	if err != nil {
		// If w is nil, the op was synchronous.
		return err
	}
	return tpgresource.OperationWait(w, activity, timeout, config.PollInterval)
}
