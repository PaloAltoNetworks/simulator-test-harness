// Package common implements the utility functions shared across repositories.
package common

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Log is a new logrus logger.
var Log = log.New()

const (
	// defaultCharset is the charset from which random strings are generated.
	defaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// The initialization function for any common shared functionality.
func init() {

	// Use full timestamps in the logs
	Log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

}

// Execute provides a wrapper to run a command with the arguments specified in args,
// in the environment defined by os.Environ, augmented with the values provided in
// the env map. The working directory of the command can also be specified in the
// workDir argument, if not specified the current working directory will be used.
// The silent argument is used for sending the output in both the variables that
// are used internally as well as the OS stdout and stderr.
func Execute(env map[string]string, command string, workDir string,
	silent bool, args ...string) (string, string, error) {

	Log.WithFields(log.Fields{
		"env":     env,
		"workdir": workDir,
		"silent":  silent,
		"args":    args,
	}).Debugf("execute command: %s", command)

	// Save the command for execution.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	//cmd := exec.Command(command, args...)

	// Execute on standard environment and any environment variable provided by the user.
	cmd.Env = os.Environ()

	// Set the command working directory.
	cmd.Dir = workDir

	// Add the values provided in env in the environment variables that will be used
	// when executing the command.
	for key, val := range env {
		cmd.Env = append(cmd.Env, key+"="+val)
	}

	var outBuf, errBuf bytes.Buffer

	// By default we output everything in both the os stdout and stderr
	// as well as our internal variables.
	cmd.Stdout = io.MultiWriter(os.Stdout, &outBuf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &errBuf)
	cmd.Stdin = os.Stdin

	// When the silent flag is true, then we do not output anything in
	// the default os stdout and stderr.
	if silent {
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
		cmd.Stdin = nil
	}

	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("start command %q: %v", command, err)
	}

	if err := cmd.Wait(); err != nil {
		return outBuf.String(), errBuf.String(),
			fmt.Errorf("wait for command %q: %v", command, err)
	}

	return outBuf.String(), errBuf.String(), nil
}

// Roulette generates a random integer in [0,max). It panics if max <= 0.
func Roulette(max int) int {

	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	src := rand.NewSource(time.Now().UnixNano())
	rgen := rand.New(src)
	num := rgen.Intn(max)

	return num
}

// RandomString generates a random string of a specified length.
func RandomString(length uint) string {

	b := make([]byte, length)
	max := len(defaultCharset)

	for i := range b {
		b[i] = defaultCharset[Roulette(max)]
	}

	return string(b)
}

// CopyFile copies src to dst. dst will be overwritten if it exists.
func CopyFile(src, dst string) error {

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %v", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create %s: %v", dst, err)
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("copy contents: %v", err)
	}

	inInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat file: %v", err)
	}
	err = os.Chmod(dst, inInfo.Mode())
	if err != nil {
		return fmt.Errorf("setting file permissions: %v", err)
	}

	return nil
}

// Stop returns a pointer to a string with the same contents as s. Especially useful when in desire
// to use a pointer to a string literal (which is impossible in Go).
func Stop(s string) *string {

	return &s
}

// ParseYamlFile is a simple function to parse a yaml/json file and unpack into
// a struct.
func ParseYamlFile(path string, i interface{}) error {

	config, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("getting config file contents: %v", err)
	}

	err = yaml.Unmarshal(config, i)
	if err != nil {
		return fmt.Errorf("unmarshaling %s: %v", path, err)
	}

	return nil
}

// RandName generates a name of words with given length separated by a delimiter.
// Format is <word1><delimiter><word2>.
func RandName(words int, wordlen int, delim string) string {

	name := RandomString(uint(wordlen))

	for i := 1; i < words; i++ {
		name = fmt.Sprintf("%s%s%s", name, delim, RandomString(uint(wordlen)))
	}

	return name
}

// RandIP generates a valid random IP
func RandIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", Roulette(250)+1, Roulette(252), Roulette(253), Roulette(254))
}
