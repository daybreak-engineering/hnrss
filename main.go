package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

var (
	bindAddr    = flag.String("bind", envOr("HNRSS_BIND", "127.0.0.1:9000"), "HOST:PORT (env HNRSS_BIND)")
	siteURL     = flag.String("site-url", envOr("HNRSS_SITE_URL", "https://hnrss.org"), "public base URL for feed self links (env HNRSS_SITE_URL)")
	buildString string
)

func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func registerEndpoint(r *gin.Engine, url string, fn gin.HandlerFunc) {
	r.GET(url, SetFormat("rss"), fn)
	r.GET(url+".jsonfeed", SetFormat("jsonfeed"), fn)
	r.GET(url+".atom", SetFormat("atom"), fn)
}

func main() {
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	registerEndpoint(r, "/newest", Newest)
	registerEndpoint(r, "/frontpage", Frontpage)
	registerEndpoint(r, "/newcomments", Newcomments)
	registerEndpoint(r, "/ask", AskHN)
	registerEndpoint(r, "/show", ShowHN)
	registerEndpoint(r, "/polls", Polls)
	registerEndpoint(r, "/jobs", Jobs)
	registerEndpoint(r, "/user", UserAll)
	registerEndpoint(r, "/threads", UserThreads)
	registerEndpoint(r, "/submitted", UserSubmitted)
	registerEndpoint(r, "/replies", Replies)
	registerEndpoint(r, "/item", Item)
	registerEndpoint(r, "/favorites", Favorites)
	registerEndpoint(r, "/bestcomments", BestComments)
	registerEndpoint(r, "/classic", Special("https://news.ycombinator.com/classic", "Hacker News: Classic"))
	registerEndpoint(r, "/best", Special("https://news.ycombinator.com/best", "Hacker News: Best"))
	registerEndpoint(r, "/invited", Special("https://news.ycombinator.com/invited", "Hacker News: Invited"))
	registerEndpoint(r, "/pool", Special("https://news.ycombinator.com/pool", "Hacker News: Pool"))
	registerEndpoint(r, "/active", Special("https://news.ycombinator.com/active", "Hacker News: Active"))
	registerEndpoint(r, "/launches", Special("https://news.ycombinator.com/launches", "Hacker News: Launches"))
	registerEndpoint(r, "/whoishiring/jobs", SeekingEmployees)
	registerEndpoint(r, "/whoishiring/hired", SeekingEmployers)
	registerEndpoint(r, "/whoishiring/freelance", SeekingFreelance)
	registerEndpoint(r, "/whoishiring", SeekingAll)

	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "https://news.ycombinator.com/favicon.ico")
	})
	r.GET("/robots.txt", func(c *gin.Context) {
		c.String(http.StatusOK, "User-agent: *\nDisallow:\n")
	})
	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok\n")
	})
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "https://hnrss.github.io/")
	})

	flag.Parse()

	srv := &http.Server{
		Addr:    *bindAddr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Kubernetes stops pods with SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown: %s\n", err)
	}
	log.Println("Server exiting")
}
