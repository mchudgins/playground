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

package gitWrapper

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	gitlab "github.com/xanzy/go-gitlab"
	"go.uber.org/zap"
)

func (g *GitWrapper) CreateMergeRequestFromBranch(ctx context.Context, title, branch string) error {
	logger := g.Logger
	logger.Debug("CreateMergeRequestFromBranch+",
		zap.String("repo", g.Repository),
		zap.String("branch", branch))
	defer logger.Debug("CreateMergeRequestFromBranch-")

	if g.repoURL == nil {
		var err error
		g.repoURL, err = url.Parse(g.Repository)
		if err != nil {
			logger.Error("Unable to parse repository as URL",
				zap.String("repository", g.Repository),
				zap.Error(err))
			return err
		}
	}

	logger.Debug("url parsed", zap.String("hostname", g.repoURL.Hostname()))

	suffix := strings.LastIndex(g.repoURL.Path, ".git")
	c := gitlab.NewClient(&http.Client{}, g.GitlabToken)
	p, _, err := c.Projects.GetProject(g.repoURL.Path[1:suffix], gitlab.WithContext(ctx))
	if err != nil {
		logger.Error("while getting project information",
			zap.Error(err),
			zap.String("project", g.repoURL.Path[1:suffix]))
	}
	logger.Debug("project info obtained",
		zap.String("project", p.Name),
		zap.Int("id", p.ID))

	_, response, err := c.MergeRequests.CreateMergeRequest(p.ID,
		&gitlab.CreateMergeRequestOptions{
			Title:           gitlab.String(title),
			SourceBranch:    gitlab.String(branch),
			TargetBranch:    gitlab.String("master"),
			TargetProjectID: gitlab.Int(p.ID),
		}, gitlab.WithContext(ctx))

	if err != nil {
		var b []byte
		if response != nil {
			b, _ = ioutil.ReadAll(response.Body)
		}
		logger.Error("unable to create merge request",
			zap.Error(err),
			zap.ByteString("body", b))
		return err
	}

	return nil
}
