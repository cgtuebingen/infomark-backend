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
  "context"
  "encoding/json"
  "fmt"
  "net/http"
  "testing"
  "time"

  "github.com/cgtuebingen/infomark-backend/api/helper"
  "github.com/cgtuebingen/infomark-backend/email"
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/franela/goblin"
  "github.com/spf13/viper"
)

func SetSheetContext(sheet *model.Sheet) func(http.Handler) http.Handler {
  return func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      ctx := context.WithValue(r.Context(), "sheet", sheet)
      next.ServeHTTP(w, r.WithContext(ctx))
    })
  }
}

func TestSheet(t *testing.T) {
  g := goblin.Goblin(t)
  email.DefaultMail = email.VoidMail

  tape := &Tape{}

  var stores *Stores

  g.Describe("Sheet", func() {

    g.BeforeEach(func() {
      tape.BeforeEach()
      stores = NewStores(tape.DB)
      _ = stores
    })

    g.It("Query should require access claims", func() {

      w := tape.Play("GET", "/api/v1/courses/1/sheets")
      g.Assert(w.Code).Equal(http.StatusUnauthorized)

      w = tape.PlayWithClaims("GET", "/api/v1/courses/1/sheets", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)
    })

    g.It("Should list all sheets a course", func() {

      w := tape.PlayWithClaims("GET", "/api/v1/courses/1/sheets", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      sheets_actual := []model.Sheet{}
      err := json.NewDecoder(w.Body).Decode(&sheets_actual)
      g.Assert(err).Equal(nil)
      g.Assert(len(sheets_actual)).Equal(10)
    })

    g.It("Should get a specific sheet", func() {

      sheet_expected, err := stores.Sheet.Get(1)
      g.Assert(err).Equal(nil)

      w := tape.PlayWithClaims("GET", "/api/v1/sheets/1", 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      sheet_actual := &model.Sheet{}
      err = json.NewDecoder(w.Body).Decode(sheet_actual)
      g.Assert(err).Equal(nil)

      g.Assert(sheet_actual.ID).Equal(sheet_expected.ID)
      g.Assert(sheet_actual.Name).Equal(sheet_expected.Name)
      g.Assert(sheet_actual.PublishAt.Equal(sheet_expected.PublishAt)).Equal(true)
      g.Assert(sheet_actual.DueAt.Equal(sheet_expected.DueAt)).Equal(true)
    })

    g.It("Creating a sheet should require access claims", func() {
      w := tape.PlayData("POST", "/api/v1/courses/1/sheets", H{})
      g.Assert(w.Code).Equal(http.StatusUnauthorized)
    })

    g.It("Creating a sheet should require access body", func() {
      w := tape.PlayDataWithClaims("POST", "/api/v1/courses/1/sheets", H{}, 1, true)
      g.Assert(w.Code).Equal(http.StatusBadRequest)
    })

    g.It("Should create valid sheet", func() {
      course_active, err := stores.Course.Get(1)
      g.Assert(err).Equal(nil)

      sheets_before, err := stores.Sheet.SheetsOfCourse(course_active, false)
      g.Assert(err).Equal(nil)

      sheet_sent := model.Sheet{
        Name:      "Sheet_new",
        PublishAt: helper.Time(time.Now()),
        DueAt:     helper.Time(time.Now()),
      }

      w := tape.PlayDataWithClaims("POST", "/api/v1/courses/1/sheets",
        tape.ToH(sheet_sent), 1, true)
      g.Assert(w.Code).Equal(http.StatusCreated)

      sheet_return := &model.Sheet{}
      err = json.NewDecoder(w.Body).Decode(&sheet_return)
      g.Assert(sheet_return.Name).Equal("Sheet_new")
      g.Assert(sheet_return.PublishAt.Equal(sheet_sent.PublishAt)).Equal(true)
      g.Assert(sheet_return.DueAt.Equal(sheet_sent.DueAt)).Equal(true)

      sheets_after, err := stores.Sheet.SheetsOfCourse(course_active, false)
      g.Assert(err).Equal(nil)
      g.Assert(len(sheets_after)).Equal(len(sheets_before) + 1)
    })

    g.It("Should skip non-existent sheet file", func() {
      w := tape.PlayWithClaims("GET", "/api/v1/sheets/1/file", 1, true)
      g.Assert(w.Code).Equal(http.StatusNotFound)

      // sheet_active, err := stores.Sheet.Get(1)
      // g.Assert(err).Equal(nil)
    })

    g.It("Should upload sheet file", func() {

      defer helper.NewSheetFileHandle(1).Delete()
      g.Assert(helper.NewSheetFileHandle(1).Exists()).Equal(false)

      // no file so far
      sheet_active, err := stores.Sheet.Get(1)
      g.Assert(err).Equal(nil)

      // upload file
      filename := fmt.Sprintf("%s/empty.zip", viper.GetString("fixtures_dir"))
      body, ct, err := tape.CreateFileRequestBody(filename, "image/jpg")
      g.Assert(err).Equal(nil)

      r, _ := http.NewRequest("POST", "/api/v1/sheets/1/file", body)
      r.Header.Set("Content-Type", ct)
      w := tape.PlayRequestWithClaims(r, 1, true)
      g.Assert(w.Code).Equal(http.StatusOK)

      // check disk
      g.Assert(helper.NewSheetFileHandle(1).Exists()).Equal(true)

      _ = sheet_active
    })

    g.AfterEach(func() {
      tape.AfterEach()
    })
  })

}
