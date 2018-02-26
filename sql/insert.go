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
	project string
	page    string
	hits    int
	size    int
}

var (
	records []Record
	logger  *zap.Logger
)

func AppendRecord(project, page, hits, size string) {
	logger.Debug("append",
		zap.String("project", project),
		zap.String("page", page),
		zap.String("hits", hits),
		zap.String("size", size))
}

func (cmd *SQL) InsertFile(ctx context.Context, filename string, limit int, dsn string) error {
	cmd.Logger.Debug("InsertFile+")
	defer cmd.Logger.Debug("InsertFile-")
	logger = cmd.Logger

	records, err := parseFile(filename)
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
					zap.String("project", records[i].project),
					zap.String("page", records[i].page),
					zap.Int("hits", records[i].hits),
					zap.Int("size", records[i].size),
					zap.Int("i", i))
			*/

			// aggregate multiple inserts into a single statement in order to reduce network traffic
			_, err = db.ExecContext(ctx,
				"insert into hits (project, page, hits, size) values (?, ?, ?, ?), (?, ?, ?, ?), "+
					"(?, ?, ?, ?), (?, ?, ?, ?), "+
					"(?, ?, ?, ?), (?, ?, ?, ?), "+
					"(?, ?, ?, ?), (?, ?, ?, ?), "+
					"(?, ?, ?, ?), (?, ?, ?, ?);",
				records[i].project, records[i].page, records[i].hits, records[i].size,
				records[i+1].project, records[i+1].page, records[i+1].hits, records[i+1].size,
				records[i+2].project, records[i+2].page, records[i+2].hits, records[i+2].size,
				records[i+3].project, records[i+3].page, records[i+3].hits, records[i+3].size,
				records[i+4].project, records[i+4].page, records[i+4].hits, records[i+4].size,
				records[i+5].project, records[i+5].page, records[i+5].hits, records[i+5].size,
				records[i+6].project, records[i+6].page, records[i+6].hits, records[i+6].size,
				records[i+7].project, records[i+7].page, records[i+7].hits, records[i+7].size,
				records[i+8].project, records[i+8].page, records[i+8].hits, records[i+8].size,
				records[i+9].project, records[i+9].page, records[i+9].hits, records[i+9].size)
			if err != nil {
				cmd.Logger.Error("Insert Error", zap.Error(err),
					zap.Int("i", i),
					zap.Int("size[i]", records[i].size),
					zap.Int("size[i+1]", records[i+1].size))
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
