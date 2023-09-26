# Overview
`gomon-client` is an integration which connects your Go project with the [gomon](https://github.com/jdudmesh/gomon) hot reload tool. 

The `gomon` tool runs your project (`go run`) and watches for file changes. It does a hard reload (process restart) on configured file extensions e.g. `.go` or soft reload (either template reload or generic callback) on an alternative set of file extensions.

`gomon` will either send your project the `SIGTERM` signal to terminate the process (it will `SIGKILL` the process if it doesn't shut down cleanly) for a hard reload or send an IPC event to trigger template reloading.

# Sample Usage
Example setup for a Labstack Echo project:

```go
package main

import (
	"fmt"
	"net/http"
	"os"

	templates "github.com/jdudmesh/gomon/pkg/client"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.Static("/assets", "./static")

	t, err := templates.NewEcho("views/*.html", e.Logger)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer t.Close()
	if err := t.Run(); err != nil {
		panic(err)
	}

	e.Renderer = t

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "index.html", nil)
	})

	if p, ok := os.LookupEnv("PORT"); ok {
		e.Logger.Fatal(e.Start(":" + p))
	} else {
		e.Logger.Fatal(e.Start(":8080"))
	}
}
```

Full code for this example can be found at [https://github.com/jdudmesh/gomon-example]

A generic `gomon` client can be implemented by creating a new instance of the ReloadManager and supplying values for the reloader client and logger parameters.

```go
package main

import (
	"fmt"
	"net/http"
	"os"

	templates "github.com/jdudmesh/gomon/pkg/client"
	log "github.com/sirupsen/logrus"
)

type app struct {}

func (a *app) Reload(filename string) {
  // do some work
}

func main() {
  a := new(app)
  logger := log.New()
  reloader = templates.New(a, logger)
}
```