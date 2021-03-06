// Copyright © 2017 Mike Hudgins <mchudgins@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package sql

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

type Record struct {
	Project string
	Page    string
	Hits    int
	Size    int
}

var (
	records []Record
)

func (cmd *SQL) InsertFile(ctx context.Context, filename string, limit int, dsn string) error {
	cmd.Logger.Debug("InsertFile+")
	defer cmd.Logger.Debug("InsertFile-")

	records, err := ParseFile(filename)
	if err != nil {
		return err
	}
	cmd.Logger.Debug("file parsed", zap.Int("records parsed", len(records)))

	recordLen := len(records)
	if limit > recordLen {
		limit = recordLen
	}

	var subset int
	var threads int = 8

	subset = limit / threads
	type response struct {
		err error
		id  int
	}
	completion := make(chan response)

	insert := func(id, start, end int, c chan response) {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			c <- response{err: err, id: id}
		}
		defer func() {
			err = db.Close()
		}()

		for i := start; i < end; i = i + 10 {
			var j int
			j = i / 10000
			if j*10000 == i {
				cmd.Logger.Debug("re-opening mysql connection",
					zap.Int("i", i),
					zap.Int("id", id))
				err = db.Close()
				if err != nil {
					cmd.Logger.Error("while closing db connection", zap.Error(err))
				}
				db, err = sql.Open("mysql", dsn)
				if err != nil {
					c <- response{err: err, id: id}
				}
			}
			/*
				cmd.Logger.Debug("record",
					zap.String("project", records[i].Project),
					zap.String("page", records[i].Page),
					zap.Int("hits", records[i].Hits),
					zap.Int("size", records[i].Size),
					zap.Int("i", i))
			*/

			// aggregate multiple inserts into a single statement in order to reduce network traffic
			_, err = db.ExecContext(ctx,
				"insert into hits (project, page, hits, size) values (?, ?, ?, ?), (?, ?, ?, ?), "+
					"(?, ?, ?, ?), (?, ?, ?, ?), "+
					"(?, ?, ?, ?), (?, ?, ?, ?), "+
					"(?, ?, ?, ?), (?, ?, ?, ?), "+
					"(?, ?, ?, ?), (?, ?, ?, ?);",
				records[i].Project, records[i].Page, records[i].Hits, records[i].Size,
				records[i+1].Project, records[i+1].Page, records[i+1].Hits, records[i+1].Size,
				records[i+2].Project, records[i+2].Page, records[i+2].Hits, records[i+2].Size,
				records[i+3].Project, records[i+3].Page, records[i+3].Hits, records[i+3].Size,
				records[i+4].Project, records[i+4].Page, records[i+4].Hits, records[i+4].Size,
				records[i+5].Project, records[i+5].Page, records[i+5].Hits, records[i+5].Size,
				records[i+6].Project, records[i+6].Page, records[i+6].Hits, records[i+6].Size,
				records[i+7].Project, records[i+7].Page, records[i+7].Hits, records[i+7].Size,
				records[i+8].Project, records[i+8].Page, records[i+8].Hits, records[i+8].Size,
				records[i+9].Project, records[i+9].Page, records[i+9].Hits, records[i+9].Size)
			if err != nil {
				cmd.Logger.Error("Insert Error", zap.Error(err),
					zap.Int("i", i),
					zap.Int("size[i]", records[i].Size),
					zap.Int("size[i+1]", records[i+1].Size))
			}
		}

		c <- response{err: nil, id: id}
	}

	for i := 0; i < threads; i++ {
		go insert(i, subset*i, subset*(i+1), completion)
	}

	for i := 0; i < threads-1; i++ {
		res := <-completion
		cmd.Logger.Debug("thread completion event", zap.Int("thread", res.id), zap.Error(res.err))

		if res.err != nil {
			return err
		}
	}

	res := <-completion
	cmd.Logger.Debug("thread completion event", zap.Int("thread", res.id), zap.Error(res.err))
	return res.err
}
