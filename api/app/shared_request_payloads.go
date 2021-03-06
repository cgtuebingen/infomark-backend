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
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
)

// EmailRequest is the request payload containing email information.
type EmailRequest struct {
	Subject string `json:"subject" example:"Switch to another day"`
	Body    string `json:"body" example:"Xmax will be from now on on 26th of Nov."`
}

// Bind preprocesses a userRequest.
func (body *EmailRequest) Bind(r *http.Request) error {

	err := validation.ValidateStruct(body,
		validation.Field(&body.Body, validation.Required),
		validation.Field(&body.Subject, validation.Required),
	)

	return err
}
