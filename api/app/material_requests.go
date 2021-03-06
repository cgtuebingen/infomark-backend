// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  ComputerGraphics Tuebingen
// Authors: Patrick Wieschollek
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package app

import (
	"errors"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
)

// MaterialRequest is the request payload for Material management.
type MaterialRequest struct {
	Name         string    `json:"name" example:"Einfuehrung"`
	Kind         int       `json:"kind" example:"1"`
	PublishAt    time.Time `json:"publish_at" example:"auto"`
	LectureAt    time.Time `json:"lecture_at" example:"auto"`
	RequiredRole int       `json:"required_role" example:"1"`
}

// Bind preprocesses a MaterialRequest.
func (body *MaterialRequest) Bind(r *http.Request) error {

	if body == nil {
		return errors.New("missing \"materials\" data")
	}

	return body.Validate()
}

// Validate validates a `Material` object.
func (body *MaterialRequest) Validate() error {
	return validation.ValidateStruct(body,
		validation.Field(
			&body.Name,
			validation.Required,
		),
		validation.Field(
			&body.PublishAt,
			validation.Required,
		),
		validation.Field(
			&body.LectureAt,
			validation.Required,
		),
		validation.Field(
			&body.Kind,
			// validation.Required,
			validation.Min(0),
			validation.Max(1),
		),
		validation.Field(
			&body.RequiredRole,
			// validation.Required,
			validation.Min(0),
			validation.Max(2),
		),
	)
}
