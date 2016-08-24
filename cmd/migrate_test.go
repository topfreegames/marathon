package cmd_test

//
// import (
// 	"os/exec"
//
// 	"git.topfreegames.com/topfreegames/marathon/cmd"
// 	mt "git.topfreegames.com/topfreegames/marathon/testing"
// 	. "github.com/onsi/ginkgo"
// 	. "github.com/onsi/gomega"
// 	"github.com/uber-go/zap"
// )
//
// var _ = Describe("Migrations", func() {
// 	var (
// 		l zap.Logger
// 	)
// 	BeforeEach(func() {
// 		l = mt.NewMockLogger()
// 		cmd := exec.Cmd{
// 			Dir:  "../",
// 			Path: "/usr/bin/make",
// 			Args: []string{
// 				"drop-test",
// 			},
// 		}
// 		_, err := cmd.CombinedOutput()
// 		Expect(err).NotTo(HaveOccurred())
// 	})
//
// 	Describe("Should migrate", func() {
// 		It("Should run migrations up", func() {
// 			cmd.InitConfig(l, "./../config/test.yaml")
// 			err := cmd.RunMigrations("./../db/migrations", -1, l)
// 			Expect(err).NotTo(HaveOccurred())
// 		})
// 	})
// })
