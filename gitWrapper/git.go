// Copyright © 2018 Mike Hudgins <mchudgins@gmail.com>
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
	"os"
	"time"

	"fmt"

	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

const (
	certPath       string = "certificates/"
	gitlabUsername string = "Certificate Management Bot"
	email          string = "certificateManagement@dstcorp.io"
)

type GitWrapper struct {
	Logger         *zap.Logger
	Repository     string
	GitlabPassword string
}

func (g *GitWrapper) AddOrUpdateFile(ctx context.Context, certname string, alternatives []string, requestor, cert string) error {
	g.Logger.Debug("AddOrUpdateFile+", zap.String("certname", certname))
	defer g.Logger.Debug("AddOrUpdateFile-")

	filename := certPath + certname + ".pem"

	// clone the repo locally
	tmpDir, err := ioutil.TempDir("", "s")
	if err != nil {
		return err
	}
	defer os.Remove(tmpDir)

	g.Logger.Debug("cloning repo",
		zap.String("repo", g.Repository),
		zap.String("tmpdir", tmpDir))

	basicAuth := &http.BasicAuth{
		Username: "dst_certificate_management",
		Password: "",
	}

	r, err := git.PlainCloneContext(ctx, tmpDir, false, &git.CloneOptions{
		URL:  g.Repository,
		Auth: basicAuth,
	})
	if err != nil {
		g.Logger.Error("unable to clone repo",
			zap.Error(err),
			zap.String("repo", g.Repository))
		return err
	}

	wt, err := r.Worktree()
	if err != nil {
		return err
	}

	// add the certificate
	err = g.writeFile(ctx, tmpDir+"/"+filename, cert)
	if err != nil {
		return err
	}

	// commit the changes to the local repo with suitable comments
	addCommit, err := wt.Add(filename)
	if err != nil {
		return err
	}
	g.Logger.Debug("add successful", zap.Any("addCommit", addCommit), zap.String("filename", filename))

	opts := &git.CommitOptions{
		All: false,
		Author: &object.Signature{
			Name:  gitlabUsername,
			Email: email,
			When:  time.Now(),
		},
	}
	commitMsg := fmt.Sprintf("Committing vault generated certificate: %s\n\nSubject Alternative Names:  %+v\nRequestor: %s\n",
		certname, alternatives, requestor)
	commit, err := wt.Commit(commitMsg, opts)
	if err != nil {
		return err
	}
	g.Logger.Debug("commit successful", zap.Any("commit", commit))

	// push the changes to the remote repository
	branch := fmt.Sprintf("certificates/%s-%d", certname, time.Now().Unix())
	refSpecs := make([]config.RefSpec, 1)
	refSpecs[0] = config.RefSpec("refs/heads/master:refs/heads/" + branch)
	pushOpts := &git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   refSpecs,
		Auth:       basicAuth,
	}
	err = r.PushContext(ctx, pushOpts)
	if err != nil {
		return err
	}

	// create a merge request so somebody will merge it into the master branch

	return err
}

func (g *GitWrapper) writeFile(ctx context.Context, filename, cert string) error {
	g.Logger.Debug("writeFile+")
	defer g.Logger.Debug("writeFile-")

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			g.Logger.Warn("unable to close file", zap.String("filename", filename), zap.Error(err))
		}
	}()

	file.WriteString(cert)

	return err
}
