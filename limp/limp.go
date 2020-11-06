// SPDX-License-Identifier: BSD-3-Clause
//
// Authors: Alexander Jung <alex@nderjung.net>
//
// Copyright (c) 2020, Alexander Jung.  All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. Neither the name of the copyright holder nor the names of its
//    contributors may be used to endorse or promote products derived from
//    this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.
package limp

import (
	"io"
	"os"
	"fmt"
	"bytes"
	"bufio"

	"github.com/hpcloud/tail"
)

// Options
type Options struct {
	InputFile  string
	OutputFile string
	TailLength int
}

var NEWLINE = []byte{'\n'}

// Limp parses the options and begins the operation of Limp
func Limp(o Options) error {
	if o.OutputFile == "" {
		return fmt.Errorf("output file not specified")
	}

	out, err := os.OpenFile(o.OutputFile, os.O_RDWR|os.O_CREATE, 0655)
	if err != nil {
		return err
	}

	defer out.Close()

	cur, err := lineCounter(out)
	if err != nil {
		return err
	}

	if o.InputFile == "" {
		err = teeStdin(out, cur, o)
	} else {
		err = teeFile(out, cur, o)
	}

	return err
}

// teeStdin reads lines from stdin and passes it to Limp
func teeStdin(out *os.File, cur int, o Options) error {
	var err error
	var line string

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line = scanner.Text()
		cur, err = limp(out, line, o.TailLength, cur)
		fmt.Printf(line + string(NEWLINE))
		if err != nil {
			return err
		}
	}

	return nil
}

// teeFile follows new lines from a file and passes it to Limp
func teeFile(out *os.File, cur int, o Options) error {
	_, err := os.Stat(o.InputFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("no such input file: %s", o.InputFile)
	}

	t, err := tail.TailFile(o.InputFile, tail.Config{ Follow: true, })
	if err != nil {
		return err
	}

	for line := range t.Lines {
		cur, err = limp(out, line.Text, o.TailLength, cur)
		if err != nil {
			return err
		}
	}

	return nil
}

// limp reads in a line and removes the head line if the length is matched
func limp(out *os.File, line string, max int, cur int) (int, error) {
	var pop int = 0

	// Truncate file if already too big
	if cur >= max {
		pop = cur - max
	}
	
	if pop > 0 {
		_, err := popLines(out, cur - max)
		if err != nil {
			return max, err
		}
		cur -= cur - max
	}
	
	cur += 1
	
	return cur, appendLine(out, line)
}

// lineCounter reads the number of lines in a file
func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], NEWLINE)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

// appendLine, well, appends a line to a file
func appendLine(f *os.File, line string) error {
	_, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	_, err = f.WriteString(line + string(NEWLINE))
	return err
}

// popLines removes n number of lines form the top of a file
func popLines(f *os.File, n int) ([]byte, error) {
	fi, err := f.Stat()

	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(make([]byte, 0, fi.Size()))

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(buf, f)
	if err != nil {
		return nil, err
	}

	lines := []byte{}

	// read n number of lines into buffer
	for n > 0 {
		l, err := buf.ReadBytes('\n')

		if err != nil && err != io.EOF {
			return nil, err
		}

		lines = append(lines, l...)
		n--
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	nw, err := io.Copy(f, buf)
	if err != nil {
		return nil, err
	}

	err = f.Truncate(nw)
	if err != nil {
		return nil, err
	}

	err = f.Sync()
	if err != nil {
		return nil, err
	}

	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return lines, nil
}
