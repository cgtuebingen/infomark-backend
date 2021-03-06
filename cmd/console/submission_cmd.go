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

package console

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/cgtuebingen/infomark-backend/api/helper"
	"github.com/cgtuebingen/infomark-backend/api/shared"
	"github.com/cgtuebingen/infomark-backend/auth/authenticate"
	"github.com/cgtuebingen/infomark-backend/model"
	"github.com/cgtuebingen/infomark-backend/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	SubmissionCmd.AddCommand(SubmissionTriggerCmd)
	SubmissionCmd.AddCommand(SubmissionTriggerAllCmd)
	SubmissionCmd.AddCommand(SubmissionRunCmd)

}

var SubmissionCmd = &cobra.Command{
	Use:   "submission",
	Short: "Management of submission",
}

var SubmissionTriggerCmd = &cobra.Command{
	Use:   "trigger [submissionID]",
	Short: "put submission into testing queue",
	Long:  `will enqueue a submission again into the testing queue`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		submissionID := MustInt64Parameter(args[0], "submissionID")

		_, stores := MustConnectAndStores()

		submission, err := stores.Submission.Get(submissionID)
		failWhenSmallestWhiff(err)

		task, err := stores.Task.Get(submission.TaskID)
		failWhenSmallestWhiff(err)

		sheet, err := stores.Task.IdentifySheetOfTask(submission.TaskID)
		failWhenSmallestWhiff(err)

		course, err := stores.Sheet.IdentifyCourseOfSheet(sheet.ID)
		failWhenSmallestWhiff(err)

		grade, err := stores.Grade.GetForSubmission(submission.ID)
		failWhenSmallestWhiff(err)

		log.Println("starting producer...")

		cfg := &service.Config{
			Connection:   viper.GetString("rabbitmq_connection"),
			Exchange:     "infomark-worker-exchange",
			ExchangeType: "direct",
			Queue:        "infomark-worker-submissions",
			Key:          viper.GetString("rabbitmq_key"),
			Tag:          "SimpleSubmission",
		}

		sha256, err := helper.NewSubmissionFileHandle(submission.ID).Sha256()
		failWhenSmallestWhiff(err)

		tokenManager, err := authenticate.NewTokenAuth()
		failWhenSmallestWhiff(err)
		accessToken, err := tokenManager.CreateAccessJWT(
			authenticate.NewAccessClaims(1, true))
		failWhenSmallestWhiff(err)

		bodyPublic, err := json.Marshal(shared.NewSubmissionAMQPWorkerRequest(
			course.ID, task.ID, submission.ID, grade.ID,
			accessToken, viper.GetString("url"), task.PublicDockerImage.String, sha256, "public"))
		if err != nil {
			log.Fatalf("json.Marshal: %s", err)
		}

		bodyPrivate, err := json.Marshal(shared.NewSubmissionAMQPWorkerRequest(
			course.ID, task.ID, submission.ID, grade.ID,
			accessToken, viper.GetString("url"), task.PrivateDockerImage.String, sha256, "private"))
		if err != nil {
			log.Fatalf("json.Marshal: %s", err)
		}

		producer, _ := service.NewProducer(cfg)
		producer.Publish(bodyPublic)
		producer.Publish(bodyPrivate)

	},
}

var SubmissionRunCmd = &cobra.Command{
	Use:   "run [submissionID]",
	Short: "run tests for a submission without writing to db",
	Long:  `will enqueue a submission again into the testing queue`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		submissionID := MustInt64Parameter(args[0], "submissionID")

		_, stores := MustConnectAndStores()

		submission, err := stores.Submission.Get(submissionID)
		failWhenSmallestWhiff(err)

		task, err := stores.Task.Get(submission.TaskID)
		failWhenSmallestWhiff(err)

		log.Println("try starting docker...")

		ds := service.NewDockerService()
		defer ds.Client.Close()

		var exit int64
		var stdout string

		submissionHnd := helper.NewSubmissionFileHandle(submission.ID)
		if !submissionHnd.Exists() {
			log.Fatalf("submission file %s for id %v is missing", submissionHnd.Path(), submission.ID)
		}

		// run public test
		if task.PublicDockerImage.Valid {
			frameworkHnd := helper.NewPublicTestFileHandle(task.ID)
			if frameworkHnd.Exists() {

				log.Printf("use docker image \"%v\"\n", task.PublicDockerImage.String)
				log.Printf("use framework file \"%v\"\n", frameworkHnd.Path())
				stdout, exit, err = ds.Run(
					task.PublicDockerImage.String,
					submissionHnd.Path(),
					frameworkHnd.Path(),
					viper.GetInt64("worker_docker_memory_bytes"),
				)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Println(" --- STDOUT -- BEGIN ---")
				fmt.Println(stdout)
				fmt.Println(" --- STDOUT -- END   ---")
				fmt.Printf("exit-code: %v\n", exit)
			} else {
				fmt.Println("skip public test, there is no framework file")

			}

		} else {
			fmt.Println("skip public test, there is no docker file")
		}

		// run private test
		if task.PrivateDockerImage.Valid {
			frameworkHnd := helper.NewPrivateTestFileHandle(task.ID)
			if frameworkHnd.Exists() {

				log.Printf("use docker image \"%v\"\n", task.PrivateDockerImage.String)
				log.Printf("use framework file \"%v\"\n", frameworkHnd.Path())
				stdout, exit, err = ds.Run(
					task.PrivateDockerImage.String,
					submissionHnd.Path(),
					frameworkHnd.Path(),
					viper.GetInt64("worker_docker_memory_bytes"),
				)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Println(" --- STDOUT -- BEGIN ---")
				fmt.Println(stdout)
				fmt.Println(" --- STDOUT -- END   ---")
				fmt.Printf("exit-code: %v\n", exit)
			} else {
				fmt.Println("skip private test, there is no framework file")

			}

		} else {
			fmt.Println("skip private test, there is no docker file")
		}

	},
}

type SubmissionWithGradeID struct {
	*model.Submission
	GradeID int64 `db:"grade_id"`
}

var SubmissionTriggerAllCmd = &cobra.Command{
	Use:   "trigger_all [taskID] [kind]",
	Short: "trigger_all tests all submissions for a given task",
	Long: `Will enqueue all submissions for a given task again into the testing queue
This triggers all [kind]-tests again (private, public).
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {

		taskID := MustInt64Parameter(args[0], "taskID")

		switch args[1] {
		case "public", "private":
		default:
			log.Fatalf("kind '%s' must be one of 'public', 'private'\n", args[2])
		}

		db, stores := MustConnectAndStores()

		task, err := stores.Task.Get(taskID)
		failWhenSmallestWhiff(err)

		sheet, err := stores.Task.IdentifySheetOfTask(task.ID)
		failWhenSmallestWhiff(err)

		course, err := stores.Sheet.IdentifyCourseOfSheet(sheet.ID)
		failWhenSmallestWhiff(err)

		log.Println("starting producer...")

		cfg := &service.Config{
			Connection:   viper.GetString("rabbitmq_connection"),
			Exchange:     "infomark-worker-exchange",
			ExchangeType: "direct",
			Queue:        "infomark-worker-submissions",
			Key:          viper.GetString("rabbitmq_key"),
			Tag:          "SimpleSubmission",
		}

		submissions := []SubmissionWithGradeID{}
		err = db.Select(&submissions, `
SELECT
  s.*, g.id grade_id
FROM
  submissions s
INNER JOIN grades g ON s.id = g.submission_ID
WHERE task_id = $1
    `, task.ID)
		failWhenSmallestWhiff(err)

		producer, _ := service.NewProducer(cfg)
		logger := logrus.New()
		logger.SetFormatter(&logrus.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
		})
		logger.Out = os.Stdout

		for _, submissionWithGrade := range submissions {
			sublog := logger.WithFields(logrus.Fields{"submissionID": submissionWithGrade.ID})

			sublog.Info("Try to enqueue")

			submissionHnd := helper.NewSubmissionFileHandle(submissionWithGrade.ID)
			if !submissionHnd.Exists() {
				sublog.Warn("uploaded file does not exists --> skip")
			}

			sha256, err := helper.NewSubmissionFileHandle(submissionWithGrade.ID).Sha256()
			if err != nil {
				sublog.Warn("Skip as sha cannot be computed")
			}

			tokenManager, err := authenticate.NewTokenAuth()
			failWhenSmallestWhiff(err)
			accessToken, err := tokenManager.CreateAccessJWT(
				authenticate.NewAccessClaims(1, true))
			failWhenSmallestWhiff(err)

			var (
				body []byte
				merr error
			)

			if args[1] == "public" {
				body, merr = json.Marshal(shared.NewSubmissionAMQPWorkerRequest(
					course.ID, taskID, submissionWithGrade.ID, submissionWithGrade.GradeID,
					accessToken, viper.GetString("url"), task.PublicDockerImage.String, sha256, "public"))

			} else {
				body, merr = json.Marshal(shared.NewSubmissionAMQPWorkerRequest(
					course.ID, taskID, submissionWithGrade.ID, submissionWithGrade.GradeID,
					accessToken, viper.GetString("url"), task.PrivateDockerImage.String, sha256, "private"))
			}
			if merr != nil {
				log.Fatalf("json.Marshal: %s", merr)
			}

			producer.Publish(body)

		}

	},
}
