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

package dynamodb

import (
	"context"

	"fmt"

	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsdb "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/mchudgins/playground/sql"
	"go.uber.org/zap"
)

const (
	batchSize int = 10
)

func (cmd *DDB) UploadFile(ctx context.Context, filename string, limit int, engine string, dsn string) error {

	cmd.Logger.Debug("InsertFile+")
	defer cmd.Logger.Debug("InsertFile-")

	records, err := sql.ParseFile(filename)
	if err != nil {
		return err
	}
	cmd.Logger.Debug("file parsed", zap.Int("records parsed", len(records)))

	recordLen := len(records)
	if limit == 0 {
		limit = int(recordLen/8) * 8
	}

	if limit > recordLen {
		limit = recordLen
	}

	var subset int
	const threads int = 8

	subset = limit / threads
	type response struct {
		err error
		id  int
	}
	completion := make(chan response)

	// cassandra
	cassandraInsert := func(tid, start, end int, c chan response) {
		logger := cmd.Logger.With(zap.Int("tid", tid), zap.String("func", "cassandraInsert"))
		logger.Debug("insert+", zap.Int("start", start), zap.Int("end", end))
		defer logger.Debug("insert-")

		cluster := gocql.NewCluster("172.31.30.210", "172.31.24.131", "172.31.17.239", "172.31.31.219", "172.32.22.220")
		cluster.Keyspace = "fubar"
		cluster.Consistency = gocql.Quorum
		session, err := cluster.CreateSession()
		if err != nil {
			logger.Error("creating session", zap.Error(err))
			c <- response{err: err, id: tid}

		}
		defer session.Close()

		for i := start; i < end; i += batchSize {
			r := records[i]
			err = session.Query("INSERT INTO hits (project, page, hits, size, id) VALUES (?,?,?,?,?)",
				r.Project, r.Page, r.Hits, r.Size, uuid.New().String()).Exec()
			if err != nil {
				logger.Error("inserting record",
					zap.Error(err),
					zap.Int("index", i),
					zap.String("project", r.Project),
					zap.String("page", r.Page),
					zap.Int("hits", r.Hits),
					zap.Int("size", r.Size))
			}
		}

		c <- response{err: nil, id: tid}
	}

	// dynamodb

	session, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	//svc := awsdb.New(session)
	svc := awsdb.New(session)

	dynamoInsert := func(tid, start, end int, c chan response) {
		logger := cmd.Logger.With(zap.Int("tid", tid), zap.String("func", "dynamodbInsert"))
		logger.Debug("insert+", zap.Int("start", start), zap.Int("end", end))
		defer logger.Debug("insert-")

		type dbRecord struct {
			sql.Record
			id uuid.UUID
		}

		for i := start; i < end; i += batchSize {
			logger.Debug("outer loop", zap.Int("i", i))
			av := make([]map[string]*awsdb.AttributeValue, 0, batchSize)

			for j := i; j < i+batchSize; j++ {
				logger.Debug("inner loop", zap.Int("j", j), zap.Int("i", i))

				r := dbRecord{
					Record: records[j],
					id:     uuid.New(),
				}

				tmp, err := dynamodbattribute.MarshalMap(r)
				if err != nil {
					logger.Error("Marshalling map", zap.Error(err), zap.Int("records[] index", j))
				} else {
					av = append(av, tmp)
				}
				cmd.Logger.Debug("record",
					zap.Int("j", j),
					zap.String("record", fmt.Sprintf("%+v", r)),
					zap.String("tmp", fmt.Sprintf("%#v", tmp)))
			}

			batch := awsdb.BatchWriteItemInput{
				RequestItems: map[string][]*awsdb.WriteRequest{
					"fubar": {
						{
							PutRequest: &awsdb.PutRequest{
								Item: av[0],
							},
						},
						{
							PutRequest: &awsdb.PutRequest{
								Item: av[1],
							},
						},
						{
							PutRequest: &awsdb.PutRequest{
								Item: av[2],
							},
						},
						{
							PutRequest: &awsdb.PutRequest{
								Item: av[3],
							},
						},
						{
							PutRequest: &awsdb.PutRequest{
								Item: av[4],
							},
						},
						{
							PutRequest: &awsdb.PutRequest{
								Item: av[5],
							},
						},
						{
							PutRequest: &awsdb.PutRequest{
								Item: av[6],
							},
						},
						{
							PutRequest: &awsdb.PutRequest{
								Item: av[7],
							},
						},
						{
							PutRequest: &awsdb.PutRequest{
								Item: av[8],
							},
						},
						{
							PutRequest: &awsdb.PutRequest{
								Item: av[9],
							},
						},
					},
				},
			}
			_, err = svc.BatchWriteItemWithContext(ctx, &batch)
			if err != nil {
				cmd.Logger.Error("writing batch",
					zap.Error(err),
					zap.String("request", fmt.Sprintf("%+v", batch)))
			}
		}

		c <- response{err: nil, id: tid}
	}

	switch strings.ToLower(engine) {
	case "cassandra":
		for i := 0; i < threads; i++ {
			cmd.Logger.Debug("spinning up thread", zap.Int("tid", i))
			go cassandraInsert(i, subset*i, subset*(i+1), completion)
		}

	case "dynamodb":
		for i := 0; i < 1; i++ {
			cmd.Logger.Debug("spinning up thread", zap.Int("tid", i))
			go dynamoInsert(i, subset*i, subset*(i+1), completion)
		}
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
