// Derived from reqfleet/shibuya; modified for lib.
package token_test

import (
	"fmt"
	"net/http"
	"testing"

	authtoken "github.com/reqfleet/lib/http/auth/token"
	"github.com/stretchr/testify/assert"
)

func TestFindToken(t *testing.T) {
	validBearer := fmt.Sprintf("%s asdf", authtoken.BEARER_PREFIX)
	header := http.Header{}
	header.Add(authtoken.AuthHeader, validBearer)
	token, err := authtoken.FindBearerToken(header)
	assert.Nil(t, err)
	assert.Equal(t, "asdf", token)

	header.Del(authtoken.AuthHeader)
	inValidBearer := "b asdf"
	header.Add(authtoken.AuthHeader, inValidBearer)
	_, err = authtoken.FindBearerToken(header)
	assert.ErrorIs(t, err, authtoken.ErrInvalidToken)

	emptyBearer := ""
	header.Del(authtoken.AuthHeader)
	header.Add(authtoken.AuthHeader, emptyBearer)
	_, err = authtoken.FindBearerToken(header)
	assert.ErrorIs(t, err, authtoken.ErrEmptyToken)
}
