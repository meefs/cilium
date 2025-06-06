// Code generated by go-swagger; DO NOT EDIT.

// Copyright Authors of Cilium
// SPDX-License-Identifier: Apache-2.0

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// BgpRoutePolicyNexthopAction BGP nexthop action
//
// swagger:model BgpRoutePolicyNexthopAction
type BgpRoutePolicyNexthopAction struct {

	// Set nexthop to the IP address of itself
	Self bool `json:"self,omitempty"`

	// Don't change nexthop
	Unchanged bool `json:"unchanged,omitempty"`
}

// Validate validates this bgp route policy nexthop action
func (m *BgpRoutePolicyNexthopAction) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this bgp route policy nexthop action based on context it is used
func (m *BgpRoutePolicyNexthopAction) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *BgpRoutePolicyNexthopAction) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *BgpRoutePolicyNexthopAction) UnmarshalBinary(b []byte) error {
	var res BgpRoutePolicyNexthopAction
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
