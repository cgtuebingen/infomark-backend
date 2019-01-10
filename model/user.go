// Copyright 2019 ComputerGraphics Tuebingen. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ==============================================================================
// Authors: Patrick Wieschollek

package model

import (
	"errors"
	"regexp"
	"time"
)

// validate an email
var reEmail = regexp.MustCompile(`(?m)[^@]+@(?:student\.|)uni-tuebingen.de`)

var errUserInvalid = errors.New("Invalid User Login")

// User represents a registered user.
type User struct {
	// the id for this user.
	ID        uint       `gorm:"primary_key"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-"`

	Email          string `json:"email"` // Email is the email address for this user.
	PasswordHash   string `json:"-"`     // PasswordHash is the encrypted password.
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	StudentNumber  string `json:"student_number"`
	Specialization string `json:"specialization"`
	Term           string `json:"term"`
	Avatar         string `json:"avatar_url"`

	ResetPasswordToken  string    `json:"-"`
	ResetPasswordSentAt time.Time `json:"-"`
	ConfirmationToken   string    `json:"-"`
	ConfirmationSentAt  time.Time `json:"-"`
	ConfirmedAt         time.Time `json:"-"`

	CurrentSignInAt time.Time `json:"-"`
}

// Validate validates the required fields and formats.
func (u *User) Validate() error {
	switch {
	case len(u.FirstName) == 0:
		return errUserInvalid
	case len(u.FirstName) > 250:
		return errUserInvalid

	case len(u.LastName) == 0:
		return errUserInvalid
	case len(u.LastName) > 250:
		return errUserInvalid

	case !reEmail.MatchString(u.Email):
		return errUserInvalid
	default:
		return nil
	}
}