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

package log

import "github.com/uber-go/zap"

//CM is a Checked Message like
type CM interface {
	Write(fields ...zap.Field)
	OK() bool
}

//D is a debug logger
func D(logger zap.Logger, message string, callback ...func(l CM)) {
	log(logger, zap.DebugLevel, message, callback...)
}

//I is a info logger
func I(logger zap.Logger, message string, callback ...func(l CM)) {
	log(logger, zap.InfoLevel, message, callback...)
}

//W is a warn logger
func W(logger zap.Logger, message string, callback ...func(l CM)) {
	log(logger, zap.WarnLevel, message, callback...)
}

//E is a error logger
func E(logger zap.Logger, message string, callback ...func(l CM)) {
	log(logger, zap.ErrorLevel, message, callback...)
}

//P is a panic logger
func P(logger zap.Logger, message string, callback ...func(l CM)) {
	log(logger, zap.PanicLevel, message, callback...)
}

func defaultWrite(l CM) {
	l.Write()
}

func log(logger zap.Logger, logLevel zap.Level, message string, callback ...func(l CM)) {
	cb := defaultWrite
	if len(callback) == 1 {
		cb = callback[0]
	}
	if cm := logger.Check(logLevel, message); cm.OK() {
		cb(cm)
	}
}
