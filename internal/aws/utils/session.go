package utils

import (
	"github.com/aws/aws-sdk-go/aws/session"
)

func GetSession() *session.Session {
	return session.Must(
		session.NewSessionWithOptions(
			session.Options{
				SharedConfigState: session.SharedConfigEnable,
			},
		),
	)
}
