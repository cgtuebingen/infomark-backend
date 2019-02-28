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

package database

import (
  "github.com/cgtuebingen/infomark-backend/model"
  "github.com/jmoiron/sqlx"
)

type SheetStore struct {
  db *sqlx.DB
}

func NewSheetStore(db *sqlx.DB) *SheetStore {
  return &SheetStore{
    db: db,
  }
}

func (s *SheetStore) Get(sheetID int64) (*model.Sheet, error) {
  p := model.Sheet{ID: sheetID}
  err := s.db.Get(&p, "SELECT * FROM sheets WHERE id = $1 LIMIT 1;", p.ID)
  return &p, err
}

func (s *SheetStore) GetAll() ([]model.Sheet, error) {
  p := []model.Sheet{}
  err := s.db.Select(&p, "SELECT * FROM sheets;")
  return p, err
}

func (s *SheetStore) Create(p *model.Sheet, courseID int64) (*model.Sheet, error) {

  newID, err := Insert(s.db, "sheets", p)
  if err != nil {
    return nil, err
  }

  // get maximum order
  var maxOrder int
  err = s.db.Get(&maxOrder, "SELECT max(ordering) FROM sheet_course WHERE course_id = $1", courseID)
  if err != nil {
    return nil, err
  }

  // now associate sheet with course
  _, err = s.db.Exec(`INSERT INTO sheet_course
    (id,sheet_id,course_id,ordering)
    VALUES (DEFAULT, $1, $2, $3);`, newID, courseID, maxOrder+1)
  if err != nil {
    return nil, err
  }

  return s.Get(newID)
}

func (s *SheetStore) Update(p *model.Sheet) error {
  return Update(s.db, "sheets", p.ID, p)
}

func (s *SheetStore) Delete(sheetID int64) error {
  return Delete(s.db, "sheets", sheetID)
}

func (s *SheetStore) SheetsOfCourse(courseID int64, only_active bool) ([]model.Sheet, error) {
  p := []model.Sheet{}

  err := s.db.Select(&p, `
    SELECT
      s.id, s.created_at, s.updated_at, s.name, s.publish_at, s.due_at
    FROM
      sheet_course sc
    INNER JOIN
      courses c ON sc.course_id = c.id
    INNER JOIN
      sheets s ON sc.sheet_id = s.id
    WHERE
      sc.course_id = $1
    ORDER BY
      sc.ordering ASC;`, courseID)
  return p, err
}
