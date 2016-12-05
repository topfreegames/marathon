/*
 * Copyright (c) 2016 TFG Co <backend@tfgco.com>
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/topfreegames/marathon/worker"
	"github.com/uber-go/zap"
)

// workerCmd represents the worker command
var workerCmd = &cobra.Command{
	Use:   "start-workers",
	Short: "starts marathon workers",
	Long:  "starts marathon workers",
	Run: func(cmd *cobra.Command, args []string) {
		ll := zap.InfoLevel
		if debug {
			ll = zap.DebugLevel
		}

		l := zap.New(
			zap.NewJSONEncoder(),
			ll,
		)

		logger := l.With(
			zap.Bool("debug", debug),
			zap.Int("port", port),
			zap.String("bind", host),
		)

		logger.Debug("configuring workers...")
		w := worker.GetWorker(debug, logger)

		logger.Debug("starting worker...")
		w.Start()
	},
}

func init() {
	RootCmd.AddCommand(workerCmd)
}
