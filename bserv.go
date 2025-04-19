/*
   bserv - Simple backup server which stores uploaded files.

   Copyright (C) 2022 Vadim Kuznetsov <vimusov@gmail.com>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const pathSep = string(os.PathSeparator)

func createWorkDir(rootDir string) string {
	curTime := time.Now()
	workDir := path.Join(
		rootDir,
		fmt.Sprintf("%.4d", curTime.Year()),
		fmt.Sprintf("%.2d", curTime.Month()),
		fmt.Sprintf("%.2d", curTime.Day()),
	)
	err := os.MkdirAll(workDir, 0755)
	if err != nil && os.IsExist(err) {
		log.Fatalf("Unable to create directory, error '%v'.", err)
	}
	return workDir
}

func formatName(baseName string) string {
	curTime := time.Now()
	return fmt.Sprintf(
		"%.2d%.2d%.2d.%.9d_%s",
		curTime.Hour(),
		curTime.Minute(),
		curTime.Second(),
		curTime.Nanosecond(),
		baseName,
	)
}

func storeFile(workDir string, fileName string, reader io.ReadCloser) {
	tmpFile, err := ioutil.TempFile(workDir, "bserv-*")
	if err != nil {
		log.Fatalf("Unable to create temporary file, error '%v'.", err)
	}
	defer func() {
		err := os.Remove(tmpFile.Name())
		if err != nil && !os.IsNotExist(err) {
			log.Fatalf("Unable to remove temporary file, error '%v'.", err)
		}
	}()

	if _, err := io.Copy(tmpFile, reader); err != nil {
		log.Fatalf("Unable to save request to temporary file, error '%v'.", err)
	}
	defer func() {
		if err := tmpFile.Close(); err != nil {
			log.Fatalf("Unable to close temporary file, error '%v'.", err)
		}
	}()

	if err := tmpFile.Sync(); err != nil {
		log.Fatalf("Unable to sync temporary file, error '%v'.", err)
	}

	if err := tmpFile.Chmod(0644); err != nil {
		log.Fatalf("Unable to change mode of temporary file, error '%v'.", err)
	}

	if err := os.Rename(tmpFile.Name(), path.Join(workDir, fileName)); err != nil {
		log.Fatalf("Unable to rename temporary file, error '%v'.", err)
	}
}

func wrapHandler(rootDir string) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Wrong method.", http.StatusMethodNotAllowed)
			return
		}
		if r.Body == nil {
			http.Error(w, "Empty file body.", http.StatusBadRequest)
			return
		}
		baseName := r.URL.Query().Get("name")
		if baseName == "" {
			http.Error(w, "File name is empty.", http.StatusBadRequest)
			return
		}
		if strings.Contains(baseName, pathSep) {
			http.Error(w, "Wrong file name.", http.StatusBadRequest)
			return
		}
		workDir := createWorkDir(rootDir)
		fileName := formatName(baseName)
		storeFile(workDir, fileName, r.Body)
		defer func() {
			err := r.Body.Close()
			if err != nil {
				log.Fatalf("Unable to close request body, error '%v'.", err)
			}
		}()
	}
	return http.HandlerFunc(handler)
}

func serveRequests(rootDir string, listenOn string) {
	http.Handle("/up", wrapHandler(rootDir))
	log.Printf("Serving on '%s' in '%s'.", listenOn, rootDir)
	if err := http.ListenAndServe(listenOn, nil); err != nil {
		log.Fatalf("Unable to listen requests, error '%v'.", err)
	}
}

func parseArgs() (string, string) {
	var listenOn string
	flag.StringVar(&listenOn, "listen-on", ":2180", "Listen on address")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatalf("Usage: bserv <root-dir>")
	}
	rootDir, resolveErr := filepath.Abs(flag.Arg(0))
	if resolveErr != nil {
		log.Fatalf("Unable to resolve absolute path from '%s', error '%v'.", flag.Arg(0), resolveErr)
	}
	return rootDir, listenOn
}

func main() {
	log.SetFlags(0)
	rootDir, listenOn := parseArgs()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go serveRequests(rootDir, listenOn)
	<-signals
}
