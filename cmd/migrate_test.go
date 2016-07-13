package cmd_test

// import (
// 	"os/exec"
//
// 	"git.topfreegames.com/topfreegames/marathon/cmd"
//
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// )
//
// var _ = Describe("Migrations", func() {
// 	BeforeEach(func() {
// 		cmd := exec.Cmd{
// 			Dir:  "../",
// 			Path: "/usr/bin/make",
// 			Args: []string{
// 				"drop-test",
// 			},
// 		}
// 		_, err := cmd.CombinedOutput()
// 		Expect(err).To(BeNil())
// 	})
//
// 	Describe("Should migrate", func() {
// 		It("Should run migrations up", func() {
// 			cmd.InitConfig()
// 			err := cmd.RunMigrations("../db/migrations", -1)
// 			Expect(err).To(BeNil())
// 		})
// 	})
// })
