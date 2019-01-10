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

package store

import (
	"log"
	"strconv"

	"github.com/cgtuebingen/infomark-backend/model"
)

// GetUserFromIdString retrieves the user from the database if exists
func (ds *datastore) GetUserFromIdString(userID string) (user *model.User, err error) {

	var uid int

	user = &model.User{}

	if userID != "" {
		log.Println(userID)
		if uid, err = strconv.Atoi(userID); err == nil {
			log.Println(uid)
			err = ORM().First(&user, uid).Error
			log.Println(user)
		}
	}

	return user, err
}