package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/zpencerq/otp"
)

const layout = `
<!DOCTYPE html><html><head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<style type="text/css">body{margin:40px
auto;font-family: Arial, Helvetica, sans-serif;max-width:650px;line-height:1.6;font-size:18px;color:#444;padding:0
10px}h1,h2,h3{line-height:1.2}textarea{width: 98%%}</style>
</head>
<body>
<header><h1><a href='/'>One Time Paste</a></h1></header>
%s
</body>
</html>`

func main() {
	var store otp.OneTimeStore
	if redisUrl, ok := os.LookupEnv("REDIS_URL"); ok {
		store = otp.NewRedisStore(redisUrl)
	} else {
		store = otp.NewMemoryStore()
	}

	bots := []string{
		"Googlebot", "Yahoo!", "bingbot", "AhrefsBot", "Baiduspider", "Ezooms",
		"MJ12bot", "YandexBot", "Slackbot",
	}
	botRegex := regexp.MustCompile(fmt.Sprintf("^.*(%s).*$", strings.Join(bots, "|")))

	ttl_default := "15"
	if given, ok := os.LookupEnv("TTL_DEFAULT"); ok {
		ttl_default = given
	}

	views_default := "2"
	if given, ok := os.LookupEnv("VIEWS_DEFAULT"); ok {
		views_default = given
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w,
			layout,
			fmt.Sprintf(`<form action='/new' method='POST'>
			   <textarea cols=40 rows=20 name='content'></textarea>
			   Time to live (in minutes): <input type='text' name='ttl' value='%s' /><br />
			   Views before expiration: <input type='text' name='views' value='%s' /><br />
			   <br />
			   <input type='submit' style='float: right;' />
			 </form>`,
				ttl_default,
				views_default))
	})

	http.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			r.ParseForm()
			content := r.PostFormValue("content")
			ttl, err := strconv.Atoi(r.PostFormValue("ttl"))
			if err != nil {
				panic(err)
			}
			views, err := strconv.Atoi(r.PostFormValue("views"))
			if err != nil {
				panic(err)
			}

			otp := store.NewConn()
			defer otp.Close()
			uuid := otp.Set(content, views, 60*ttl)
			url := fmt.Sprintf("https://%s/show/%s", r.Host, uuid)
			fmt.Fprintf(w, layout,
				fmt.Sprintf(
					`<p>
					  <strong>Here's your link!</strong>
					  It'll only work %d time(s)
					</p>
					<a href="%s">%s</a>`,
					views, url, url))
		} else {
			http.Redirect(w, r, "/", 302)
			return
		}
	})

	http.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "User-agent: *\nDisallow: /")
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {})

	http.HandleFunc("/show/", func(w http.ResponseWriter, r *http.Request) {
		if botRegex.MatchString(r.Header.Get("User-Agent")) {
			http.Redirect(w, r, "/", 302)
			return
		}

		otp := store.NewConn()
		defer otp.Close()
		p := strings.Split(r.URL.Path, "/")
		if len(p) == 3 {
			key := p[2]
			if otp.Exists(key) {
				fmt.Fprintf(w, layout,
					fmt.Sprintf("<hr>%s",
						strings.Replace(
							html.EscapeString(*otp.Get(key)),
							"\n", "<br />",
							-1)),
				)
			} else {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, layout, "The paste never existed or has expired.")

			}
		} else {
			http.Redirect(w, r, "/", 302)
			return
		}
	})

	port := "8080"
	if givenPort, ok := os.LookupEnv("PORT"); ok {
		port = givenPort
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
