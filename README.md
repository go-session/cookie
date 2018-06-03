# Cookie store for [Session](https://github.com/go-session/session)

[![Build][Build-Status-Image]][Build-Status-Url] [![Coverage][Coverage-Image]][Coverage-Url] [![ReportCard][reportcard-image]][reportcard-url] [![GoDoc][godoc-image]][godoc-url] [![License][license-image]][license-url]

## Quick Start

### Download and install

```bash
$ go get -u -v gopkg.in/go-session/cookie.v1
```

### Create file `server.go`

```go
package main

import (
	"context"
	"fmt"
	"net/http"

	"gopkg.in/go-session/cookie.v1"
	"gopkg.in/session.v2"
)

var (
	hashKey = []byte("FF51A553-72FC-478B-9AEF-93D6F506DE91")
)

func main() {
	session.InitManager(
		session.SetCookieName("demo_session_id"),
		session.SetSign([]byte("sign")),
		session.SetStore(
			cookie.NewCookieStore(
				cookie.SetCookieName("demo_cookie_store_id"),
				cookie.SetHashKey(hashKey),
			),
		),
	)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		store, err := session.Start(context.Background(), w, r)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		store.Set("foo", "bar")
		err = store.Save()
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		http.Redirect(w, r, "/foo", 302)
	})

	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		store, err := session.Start(context.Background(), w, r)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		foo, ok := store.Get("foo")
		if !ok {
			fmt.Fprint(w, "does not exist")
			return
		}

		fmt.Fprintf(w, "foo:%s", foo)
	})

	http.ListenAndServe(":8080", nil)
}
```

### Build and run

```bash
$ go build server.go
$ ./server
```

### Open in your web browser

<http://localhost:8080>

    foo:bar


## MIT License

    Copyright (c) 2018 Lyric

[Build-Status-Url]: https://travis-ci.org/go-session/cookie
[Build-Status-Image]: https://travis-ci.org/go-session/cookie.svg?branch=master
[Coverage-Url]: https://coveralls.io/github/go-session/cookie?branch=master
[Coverage-Image]: https://coveralls.io/repos/github/go-session/cookie/badge.svg?branch=master
[reportcard-url]: https://goreportcard.com/report/gopkg.in/go-session/cookie.v1
[reportcard-image]: https://goreportcard.com/badge/gopkg.in/go-session/cookie.v1
[godoc-url]: https://godoc.org/gopkg.in/go-session/cookie.v1
[godoc-image]: https://godoc.org/gopkg.in/go-session/cookie.v1?status.svg
[license-url]: http://opensource.org/licenses/MIT
[license-image]: https://img.shields.io/npm/l/express.svg
