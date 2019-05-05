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
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/email"
	"github.com/franela/goblin"
)

func TestGroup(t *testing.T) {
	g := goblin.Goblin(t)
	email.DefaultMail = email.VoidMail

	tape := &Tape{}

	var stores *Stores

	g.Describe("Group", func() {

		g.BeforeEach(func() {
			tape.BeforeEach()
			stores = NewStores(tape.DB)
			_ = stores
		})

		g.It("Query should require access claims", func() {

			w := tape.Get("/api/v1/courses/1/groups")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.GetWithClaims("/api/v1/courses/1/groups", 1, true)
			g.Assert(w.Code).Equal(http.StatusOK)
		})

		g.It("Should list all groups from a course", func() {
			w := tape.GetWithClaims("/api/v1/courses/1/groups", 1, true)
			g.Assert(w.Code).Equal(http.StatusOK)

			groupsActual := []GroupResponse{}
			err := json.NewDecoder(w.Body).Decode(&groupsActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(groupsActual)).Equal(10)
		})

		g.It("Should get a specific group", func() {
			entryExpected, err := stores.Group.Get(1)
			g.Assert(err).Equal(nil)

			w := tape.GetWithClaims("/api/v1/courses/1/groups/1", 1, true)
			g.Assert(w.Code).Equal(http.StatusOK)

			entryActual := &GroupResponse{}
			err = json.NewDecoder(w.Body).Decode(entryActual)
			g.Assert(err).Equal(nil)

			g.Assert(entryActual.ID).Equal(entryExpected.ID)
			g.Assert(entryActual.Tutor.ID).Equal(entryExpected.TutorID)
			g.Assert(entryActual.CourseID).Equal(entryExpected.CourseID)
			g.Assert(entryActual.Description).Equal(entryExpected.Description)

			t, err := stores.User.Get(entryExpected.TutorID)
			g.Assert(err).Equal(nil)
			g.Assert(entryActual.Tutor.FirstName).Equal(t.FirstName)
			g.Assert(entryActual.Tutor.LastName).Equal(t.LastName)
			g.Assert(entryActual.Tutor.AvatarURL).Equal(t.AvatarURL)
			g.Assert(entryActual.Tutor.Email).Equal(t.Email)
			g.Assert(entryActual.Tutor.Language).Equal(t.Language)
		})

		g.It("Creating should require claims", func() {
			w := tape.Post("/api/v1/courses/1/groups", H{})
			g.Assert(w.Code).Equal(http.StatusUnauthorized)
		})

		g.Xit("Creating should require body", func() {
			// TODO empty request with claims
		})

		g.It("Should create valid group", func() {
			entriesBefore, err := stores.Group.GroupsOfCourse(1)
			g.Assert(err).Equal(nil)

			tutorID := int64(1)

			entrySent := helper.H{
				"tutor": helper.H{
					"id": tutorID,
				},
				"description": "blah blahe",
			}

			// err = entrySent.Validate()
			// g.Assert(err).Equal(nil)

			w := tape.PostWithClaims("/api/v1/courses/1/groups", entrySent, 1, true)
			g.Assert(w.Code).Equal(http.StatusCreated)

			entryReturn := &GroupResponse{}
			err = json.NewDecoder(w.Body).Decode(&entryReturn)
			g.Assert(err).Equal(nil)
			g.Assert(entryReturn.Tutor.ID).Equal(tutorID)
			g.Assert(entryReturn.CourseID).Equal(int64(1))
			g.Assert(entryReturn.Description).Equal("blah blahe")

			t, err := stores.User.Get(1)
			g.Assert(err).Equal(nil)
			g.Assert(entryReturn.Tutor.FirstName).Equal(t.FirstName)
			g.Assert(entryReturn.Tutor.LastName).Equal(t.LastName)
			g.Assert(entryReturn.Tutor.AvatarURL).Equal(t.AvatarURL)
			g.Assert(entryReturn.Tutor.Email).Equal(t.Email)
			g.Assert(entryReturn.Tutor.Language).Equal(t.Language)

			entriesAfter, err := stores.Group.GroupsOfCourse(1)
			g.Assert(err).Equal(nil)
			g.Assert(len(entriesAfter)).Equal(len(entriesBefore) + 1)
		})

		g.It("Should update a group", func() {
			// group (id=1) belongs to course(id=1)
			tutorID := int64(9)
			entrySent := helper.H{
				"tutor": helper.H{
					"id": tutorID,
				},
				"description": "new descr",
			}

			// students
			w := tape.PlayDataWithClaims("PUT", "/api/v1/courses/1/groups/1", entrySent, 112, false)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w = tape.PlayDataWithClaims("PUT", "/api/v1/courses/1/groups/1", entrySent, 2, false)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.PlayDataWithClaims("PUT", "/api/v1/courses/1/groups/1", entrySent, 1, false)
			g.Assert(w.Code).Equal(http.StatusOK)

			entryAfter, err := stores.Group.Get(1)
			g.Assert(err).Equal(nil)

			g.Assert(entryAfter.TutorID).Equal(tutorID)
			g.Assert(entryAfter.CourseID).Equal(int64(1))
		})

		g.It("Should delete when valid access claims", func() {
			entriesBefore, err := stores.Group.GetAll()
			g.Assert(err).Equal(nil)

			w := tape.Delete("/api/v1/courses/1/groups/1")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			// verify nothing has changes
			entriesAfter, err := stores.Group.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(entriesAfter)).Equal(len(entriesBefore))

			// students
			w = tape.DeleteWithClaims("/api/v1/courses/1/groups/1", 112, false)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutors
			w = tape.DeleteWithClaims("/api/v1/courses/1/groups/1", 2, false)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.DeleteWithClaims("/api/v1/courses/1/groups/1", 1, false)
			g.Assert(w.Code).Equal(http.StatusOK)

			// verify a sheet less exists
			entriesAfter, err = stores.Group.GetAll()
			g.Assert(err).Equal(nil)
			g.Assert(len(entriesAfter)).Equal(len(entriesBefore) - 1)
		})

		g.It("Should change a bid to a group", func() {
			userID := int64(112)

			// admins are not allowed
			w := tape.PostWithClaims("/api/v1/courses/1/groups/1/bids", H{"bid": 4}, 1, false)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			// tutors are not allowed
			w = tape.PostWithClaims("/api/v1/courses/1/groups/1/bids", H{"bid": 4}, 2, false)
			g.Assert(w.Code).Equal(http.StatusBadRequest)

			// students
			w = tape.PostWithClaims("/api/v1/courses/1/groups/1/bids", H{"bid": 4}, userID, false)
			g.Assert(w.Code).Equal(http.StatusOK)
			// no content

			// delete to test insert
			tape.DB.Exec(`DELETE FROM group_bids where user_id = $1`, userID)

			w = tape.PostWithClaims("/api/v1/courses/1/groups/1/bids", H{"bid": 4}, userID, false)
			g.Assert(w.Code).Equal(http.StatusCreated)
			entryReturn := &GroupBidResponse{}
			err := json.NewDecoder(w.Body).Decode(&entryReturn)
			g.Assert(err).Equal(nil)
			g.Assert(entryReturn.Bid).Equal(4)

		})

		g.It("Find my group when being a student", func() {
			// a random student (checked via pgweb)
			loginID := int64(112)

			w := tape.Get("/api/v1/courses/1/groups/own")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.GetWithClaims("/api/v1/courses/1/groups/own", loginID, false)
			g.Assert(w.Code).Equal(http.StatusOK)

			entryReturn := []GroupResponse{}
			err := json.NewDecoder(w.Body).Decode(&entryReturn)
			g.Assert(err).Equal(nil)

			// we cannot check the other entries
			g.Assert(entryReturn[0].CourseID).Equal(int64(1))
		})

		g.It("Find my group when being a tutor", func() {
			// a random student (checked via pgweb)
			loginID := int64(2)

			w := tape.Get("/api/v1/courses/1/groups/own")
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			w = tape.GetWithClaims("/api/v1/courses/1/groups/own", loginID, true)
			g.Assert(w.Code).Equal(http.StatusOK)

			entryReturn := []GroupResponse{}
			err := json.NewDecoder(w.Body).Decode(&entryReturn)
			g.Assert(err).Equal(nil)

			// we cannot check the other entries
			g.Assert(entryReturn[0].CourseID).Equal(int64(1))
			g.Assert(entryReturn[0].Tutor.ID).Equal(loginID)

			t, err := stores.User.Get(loginID)
			g.Assert(err).Equal(nil)
			g.Assert(entryReturn[0].Tutor.FirstName).Equal(t.FirstName)
			g.Assert(entryReturn[0].Tutor.LastName).Equal(t.LastName)
			g.Assert(entryReturn[0].Tutor.AvatarURL).Equal(t.AvatarURL)
			g.Assert(entryReturn[0].Tutor.Email).Equal(t.Email)
			g.Assert(entryReturn[0].Tutor.Language).Equal(t.Language)

		})

		g.It("Only tutors and admins can send emails to a group", func() {
			w := tape.Post("/api/v1/courses/1/groups/1/emails", H{"subject": "subj", "body": "body"})
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			// student
			w = tape.PostWithClaims("/api/v1/courses/1/groups/1/emails", H{"subject": "subj", "body": "body"}, 112, false)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutor
			w = tape.PostWithClaims("/api/v1/courses/1/groups/1/emails", H{"subject": "subj", "body": "body"}, 2, false)
			g.Assert(w.Code).Equal(http.StatusOK)

			// admin
			w = tape.PostWithClaims("/api/v1/courses/1/groups/1/emails", H{"subject": "subj", "body": "body"}, 1, false)
			g.Assert(w.Code).Equal(http.StatusOK)

		})

		g.It("Permission test", func() {
			url := "/api/v1/courses/1/groups"

			// global root can do whatever they want
			w := tape.GetWithClaims(url, 1, true)
			g.Assert(w.Code).Equal(http.StatusOK)

			// enrolled tutors can access
			w = tape.GetWithClaims(url, 2, false)
			g.Assert(w.Code).Equal(http.StatusOK)

			// enrolled students can access
			w = tape.GetWithClaims(url, 112, false)
			g.Assert(w.Code).Equal(http.StatusOK)

			// disenroll student
			w = tape.DeleteWithClaims("/api/v1/courses/1/enrollments", 112, false)
			g.Assert(w.Code).Equal(http.StatusOK)

			// cannot access anymore
			w = tape.GetWithClaims(url, 112, false)
			g.Assert(w.Code).Equal(http.StatusForbidden)
		})

		g.It("should manually add user to group (update)", func() {
			url := "/api/v1/courses/1/groups/1/enrollments"

			w := tape.Post(url, H{"user_id": 112})
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			// student
			w = tape.PostWithClaims(url, H{"user_id": 112}, 112, false)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutor
			w = tape.PostWithClaims(url, H{"user_id": 112}, 2, false)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.PostWithClaims(url, H{"user_id": 112}, 1, false)
			fmt.Println(w.Body)
			g.Assert(w.Code).Equal(http.StatusOK)

			enrollment, err := stores.Group.GetGroupEnrollmentOfUserInCourse(112, 1)
			g.Assert(err).Equal(nil)
			g.Assert(enrollment.UserID).Equal(int64(112))
			g.Assert(enrollment.GroupID).Equal(int64(1))

			// admin
			w = tape.PostWithClaims("/api/v1/courses/1/groups/2/enrollments", H{"user_id": 112}, 1, false)
			g.Assert(w.Code).Equal(http.StatusOK)

			enrollment, err = stores.Group.GetGroupEnrollmentOfUserInCourse(112, 1)
			g.Assert(err).Equal(nil)
			g.Assert(enrollment.UserID).Equal(int64(112))
			g.Assert(enrollment.GroupID).Equal(int64(2))

		})

		g.It("should manually add user to group (create)", func() {

			// remove all user_group from student
			_, err := tape.DB.Exec("DELETE FROM user_group WHERE user_id = 112;")
			g.Assert(err).Equal(nil)

			w := tape.Post("/api/v1/courses/1/groups/1/enrollments", H{"user_id": 112})
			g.Assert(w.Code).Equal(http.StatusUnauthorized)

			// student
			w = tape.PostWithClaims("/api/v1/courses/1/groups/1/enrollments", H{"user_id": 112}, 112, false)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// tutor
			w = tape.PostWithClaims("/api/v1/courses/1/groups/1/enrollments", H{"user_id": 112}, 2, false)
			g.Assert(w.Code).Equal(http.StatusForbidden)

			// admin
			w = tape.PostWithClaims("/api/v1/courses/1/groups/1/enrollments", H{"user_id": 112}, 1, false)
			g.Assert(w.Code).Equal(http.StatusOK)

			enrollment, err := stores.Group.GetGroupEnrollmentOfUserInCourse(112, 1)
			g.Assert(err).Equal(nil)
			g.Assert(enrollment.UserID).Equal(int64(112))
			g.Assert(enrollment.GroupID).Equal(int64(1))

			// admin
			w = tape.PostWithClaims("/api/v1/courses/1/groups/2/enrollments", H{"user_id": 112}, 1, false)
			g.Assert(w.Code).Equal(http.StatusOK)

			enrollment, err = stores.Group.GetGroupEnrollmentOfUserInCourse(112, 1)
			g.Assert(err).Equal(nil)
			g.Assert(enrollment.UserID).Equal(int64(112))
			g.Assert(enrollment.GroupID).Equal(int64(2))

		})

		g.It("Should be able to filter enrollments (all)", func() {
			groupActive, err := stores.Group.Get(1)
			g.Assert(err).Equal(nil)

			numberEnrollmentsExpected, err := DBGetInt(
				tape,
				"SELECT count(*) FROM user_group WHERE group_id = $1",
				groupActive.ID,
			)
			g.Assert(err).Equal(nil)

			w := tape.GetWithClaims("/api/v1/courses/1/groups/1/enrollments", 1, true)
			enrollmentsActual := []enrollmentResponse{}
			err = json.NewDecoder(w.Body).Decode(&enrollmentsActual)
			g.Assert(err).Equal(nil)
			g.Assert(len(enrollmentsActual)).Equal(numberEnrollmentsExpected)
		})

		g.AfterEach(func() {
			tape.AfterEach()
		})

	})

}
