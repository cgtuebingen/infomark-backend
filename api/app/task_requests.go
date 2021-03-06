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

	validation "github.com/go-ozzo/ozzo-validation"
)

// TaskRequest is the request payload for Task management.
type TaskRequest struct {
	MaxPoints          int    `json:"max_points" example:"25"`
	Name               string `json:"name" example:"Task 1"`
	PublicDockerImage  string `json:"public_docker_image" example:"DefaultJavaTestingImage"`
	PrivateDockerImage string `json:"private_docker_image" example:"DefaultJavaTestingImage"`
}

// Bind preprocesses a TaskRequest.
func (body *TaskRequest) Bind(r *http.Request) error {
	if body == nil {
		return errors.New("missing \"task\" data")
	}
	return body.Validate()
}

// Validate validates a TaskRequest.
func (body *TaskRequest) Validate() error {
	return validation.ValidateStruct(body,
		validation.Field(
			&body.MaxPoints,
			validation.Min(0),
		),
		validation.Field(
			&body.Name,
			validation.Required,
		),
	)
}
