// Package amzn contains AWS-specific implementation.
package amzn

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws/session"
)

var (
	// AWSSession returns an AWS session that can be used for
	// AWS operations. The calling program can override this
	// if necessary. The default implementation returns a session
	// with defaults obtained from the environment.
	AWSSession func() *session.Session
)

func init() {
	var once sync.Once
	var sess *session.Session

	AWSSession = func() *session.Session {
		once.Do(func() {
			sess = session.New()
		})
		return sess
	}
}
