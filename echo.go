package client

// gomon is a simple command line tool that watches your files and automatically restarts the application when it detects any changes in the working directory.
// Copyright (C) 2023 John Dudmesh

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

import (
	"fmt"
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type templateManagerEcho struct {
	reloadManager ReloadManager
	pathGlob      string
	templates     *template.Template
}

func NewEcho(pathGlob string) (*templateManagerEcho, error) {
	t := &templateManagerEcho{
		pathGlob: pathGlob,
	}

	t.Reload("***startup***")

	reloadManager, err := New(t.Reload)
	if err != nil {
		return nil, fmt.Errorf("initiating template manager: %w", err)
	}
	t.reloadManager = reloadManager

	return t, nil
}

func (t *templateManagerEcho) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func (t *templateManagerEcho) Reload(hint string) error {
	var err error
	t.templates, err = template.ParseGlob(t.pathGlob)
	return err
}

func (t *templateManagerEcho) Close() {
	t.reloadManager.Close()
}

func (t *templateManagerEcho) ListenAndServe() error {
	return t.reloadManager.ListenAndServe()
}
