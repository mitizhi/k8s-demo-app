package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	log "github.com/sirupsen/logrus"
	"k8s-demo-app/internal/config"
)

var (
	mu              sync.Mutex
	templateDir     string
	stateDir        string
	envMap          map[string]string
	secretPassword  string
	sensitiveInfo   string
)

type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *log.Entry) ([]byte, error) {
	// Looks weird and ad hoc, but it works:
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	level := entry.Level.String()
	message := entry.Message
	logLine := fmt.Sprintf("[%s] %-7s %s\n", timestamp, level, message)
	return []byte(logLine), nil
}

func init() {
	// Set custom log format and level
	log.SetFormatter(&CustomFormatter{})

	// Set log level according to LOGLEVEL
	logLevel := log.InfoLevel
	logLevelStr := config.GetEnvDefault("LOGLEVEL", "info")
	l := logLevelStr[0]
	switch {
	case l == 'd':
		logLevel = log.DebugLevel
	case l == 'i':
		logLevel = log.InfoLevel
	case l == 'w':
		logLevel = log.WarnLevel
	case l == 'e':
		logLevel = log.ErrorLevel
	case l == 'f':
		logLevel = log.FatalLevel
	default:
		log.Errorf("bad log level (only the first character is used): \"" + logLevelStr + "\"")
	}
	log.SetLevel(logLevel)

	secretPassword = config.GetEnvDefault("SECRETPASSWD", "secret")
	sensitiveInfo = config.GetEnvDefault("SENSITIVEINFO", "sensitive information")
	templateDir = config.GetEnvDefault("TEMPLATEDIR", "data") + "/"
	stateDir = config.GetEnvDefault("STATEDIR", "state") + "/"
	envMap = getEnvMap()
	log.Debugf("LOGLEVEL: \"" + logLevelStr + "\"")
	log.Debugf("SECRETPASSWD: \"" + secretPassword + "\" (do not tell anyone!)")
	log.Debugf("SENSITIVEINFO: \"" + sensitiveInfo + "\" (do not tell anyone!)")
	log.Debugf("TEMPLATEDIR: \"" + templateDir + "\"")
	log.Debugf("STATEDIR: \"" + stateDir + "\"")
}

func getEnvMap() map[string]string {
	envMap := make(map[string]string)
	log.Debugf("Environment variables:")
	for _, env := range os.Environ() {
		pair := split(env, '=')
		envMap[pair[0]] = pair[1]
		log.Debugf(" - \"" + pair[0] + "\" = \"" + pair[1] + "\"")
	}
	return envMap
}

func split(s string, sep rune) []string {
	for i := range s {
		if s[i] == byte(sep) {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

/**
 * @param prefix the prefix path including preceding and trailing slashes ("/").
 *               Only requests beginning with this URL path are handled.
 */
func MakeHandler(prefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path
		if !startsWithPrefix(urlPath, prefix) {
			log.Debugf("request (path \"" + urlPath + "\") does not match the prefix (prefix: \"" + prefix + "\")")
			http.NotFound(w, r)
			return
		}
		unprefixedPath := urlPath[len(prefix):]
		log.Debugf("unprefixed path: \"" + unprefixedPath + "\" - len(prefix): %d", len(prefix))

		log.Infof("request received, path: \"" + urlPath + "\" --> \"" + unprefixedPath + "\" (with prefix \"" + prefix + "\" removed)")
		switch {
		case unprefixedPath == "" || unprefixedPath =="/":
			fmt.Fprint(w, "It works!\n")
		case unprefixedPath == "version":
			fmt.Fprintf(w, "App \"" + config.GetAppName() + "\" version: " + config.GetAppVersion() + "\n")
		case unprefixedPath == "crash":
			handleCrash()
		case unprefixedPath == "quit":
			handleQuit(w, r)
		case unprefixedPath == "count":
			handleCount(w, r)
		case startsWithPrefix(unprefixedPath, "sensitive/"):
			handleSecret(w, r)
		case startsWithPrefix(unprefixedPath, "list/"):
			handleListUrls(w, r, prefix)
		default:
			handleTemplate(w, r)
		}
	}
}

func startsWithPrefix(path string, prefix string) bool {
	return len(path) >= len(prefix) && path[:len(prefix)] == prefix
}

func handleCrash() {
	log.Fatalf("oh no, the app is crashing...")
	os.Exit(1)
}

func handleQuit(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Bye, bye - preparing to lie down down and die after a second!")
	log.Infof("exiting... the container goes down in a second!")

	// Flush the response writer to ensure the client receives the response
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	go func() {
		log.Infof("Waiting 1000ms before dying")
		time.Sleep(1000 * time.Millisecond)
		log.Infof("Time to die!")
		os.Exit(0)
	}()
}

/**
 * N.B. For replicated, possibly multi-node deployment the synchronization fails.
 * A possible solution is to use an etcd service for the synchronizatrion.
 */
func handleCount(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	countFilePath := stateDir + "count"
	data, err := ioutil.ReadFile(countFilePath)
	if err != nil && !os.IsNotExist(err) {
		http.Error(w, "Error reading count file", http.StatusInternalServerError)
		return
	}

	count := 0
	if len(data) > 0 {
		count, err = strconv.Atoi(string(data))
		if err != nil {
			http.Error(w, "Error parsing count file", http.StatusInternalServerError)
			return
		}
	}

	count++
	err = ioutil.WriteFile(countFilePath, []byte(strconv.Itoa(count)), 0644)
	if err != nil {
		http.Error(w, "Error writing count file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%d\n", count)
}

func handleSecret(w http.ResponseWriter, r *http.Request) {
	providedSecret := filepath.Base(r.URL.Path)

	if providedSecret != secretPassword {
		http.Error(w, "Authentication error", http.StatusUnauthorized)
		return
	}

	fmt.Fprintf(w, "Sensitive information: \"" + sensitiveInfo + "\"\n")
}

func handleListUrls(w http.ResponseWriter, r *http.Request, prefix string) {
	host := filepath.Base(r.URL.Path)
	addUrlListEntry(w, host, prefix, "", "                    Test connectivity")
	addUrlListEntry(w, host, prefix, "crash", "               Abrupt crash scenario (status: 1)")
	addUrlListEntry(w, host, prefix, "quit", "                Exit cleanly (status: 0)")
	addUrlListEntry(w, host, prefix, "count", "               Increment counter")
	addUrlListEntry(w, host, prefix, "sensitive/<password>", "Show sensitive information")
	addUrlListEntry(w, host, prefix, "list/<host>", "         List URL schemes")
	addUrlListEntry(w, host, prefix, "${file}", "             Expand template in \"" + templateDir + "/${file}\"")
}

func addUrlListEntry(w http.ResponseWriter, host string, prefix string, path string, description string) {
	url := "http://" + host + ":" + config.GetEnvDefault("PORT", "8080") + prefix + path
	fmt.Fprintf(w, " * " + url + " " + description + "\n")
}

func handleTemplate(w http.ResponseWriter, r *http.Request) {
	filePath := templateDir + filepath.Base(r.URL.Path)
	log.Debugf("attempting to read the file \"" + filePath + "\"")
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Debugf("file not found: \"" + filePath + "\"")
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.New("").Parse(string(content))
	if err != nil {
		log.Errorf("error parsing template")
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, envMap)		// Execute the template and write to the buffer
	if err != nil {
		log.Errorf("error executing the template")
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
	log.Debug("Expanded template:\n----------\n" + tpl.String() + "----------")
	// write HTTP 200 OK response
	w.WriteHeader(http.StatusOK)
	tpl.WriteTo(w)
}

