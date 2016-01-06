// -*- Mode: Go; indent-tabs-mode: t -*-
// +build !excludeintegration

/*
 * Copyright (C) 2014-2015 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package store

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/ubuntu-core/snappy/helpers"
	"github.com/ubuntu-core/snappy/snappy"

	. "gopkg.in/check.v1"
)

// Hook up check.v1 into the "go test" runner
func Test(t *testing.T) { TestingT(t) }

type storeTestSuite struct {
	client *http.Client
	store  *Store
}

var _ = Suite(&storeTestSuite{})

func (s *storeTestSuite) SetUpTest(c *C) {
	s.store = NewStore(c.MkDir())
	err := s.store.Start()
	c.Assert(err, IsNil)

	transport := &http.Transport{}
	s.client = &http.Client{
		Transport: transport,
	}
}

func (s *storeTestSuite) TearDownTest(c *C) {
	s.client.Transport.(*http.Transport).CloseIdleConnections()
	err := s.store.Stop()
	c.Assert(err, IsNil)
}

// StoreGet gets the given from the store
func (s *storeTestSuite) StoreGet(path string) (*http.Response, error) {
	return s.client.Get(s.store.URL() + path)
}

func (s *storeTestSuite) TestStoreURL(c *C) {
	c.Assert(s.store.URL(), Equals, "http://"+defaultAddr)
}

func (s *storeTestSuite) TestTrivialGetWorks(c *C) {
	resp, err := s.StoreGet("/")
	c.Assert(err, IsNil)
	defer resp.Body.Close()

	c.Assert(resp.StatusCode, Equals, 418)
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(body), Equals, "I'm a teapot")

}

func (s *storeTestSuite) TestSearchEndpoint(c *C) {
	resp, err := s.StoreGet("/search")
	c.Assert(err, IsNil)
	defer resp.Body.Close()

	c.Assert(resp.StatusCode, Equals, 501)
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(body), Equals, "search not implemented yet")

}

func (s *storeTestSuite) TestDetailsEndpoint(c *C) {
	resp, err := s.StoreGet("/package/")
	c.Assert(err, IsNil)
	defer resp.Body.Close()

	c.Assert(resp.StatusCode, Equals, 501)
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(body), Equals, "details not implemented yet")
}

func (s *storeTestSuite) TestBulkEndpoint(c *C) {
	resp, err := s.StoreGet("/click-metadata")
	c.Assert(err, IsNil)
	defer resp.Body.Close()

	c.Assert(resp.StatusCode, Equals, 501)
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(body), Equals, "bulk not implemented yet")
}

// FIXME: extract into snappy/testutils
func (s *storeTestSuite) makeTestSnap(c *C, packageYamlContent string) string {
	tmpdir := c.MkDir()
	os.MkdirAll(filepath.Join(tmpdir, "meta"), 0755)

	packageYaml := filepath.Join(tmpdir, "meta", "package.yaml")
	ioutil.WriteFile(packageYaml, []byte(packageYamlContent), 0644)
	readmeMd := filepath.Join(tmpdir, "meta", "readme.md")

	content := "Random\nExample"
	err := ioutil.WriteFile(readmeMd, []byte(content), 0644)
	c.Assert(err, IsNil)

	targetDir := s.store.blobDir
	// FIXME: legacy? squash?
	snapFn, err := snappy.BuildLegacySnap(tmpdir, targetDir)
	c.Assert(err, IsNil)
	return snapFn
}

func (s *storeTestSuite) TestMakeTestSnap(c *C) {
	snapFn := s.makeTestSnap(c, "name: foo\nversion: 1")
	c.Assert(helpers.FileExists(snapFn), Equals, true)
	c.Assert(snapFn, Equals, filepath.Join(s.store.blobDir, "foo_1_all.snap"))
}

func (s *storeTestSuite) TestRefreshSnaps(c *C) {
	s.makeTestSnap(c, "name: foo\nversion: 1")

	s.store.refreshSnaps()
	c.Assert(s.store.snaps, DeepEquals, map[string]string{
		"foo": filepath.Join(s.store.blobDir, "foo_1_all.snap"),
	})
}
