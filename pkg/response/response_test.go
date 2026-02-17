package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-boilerplate/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("CustomError without wrap", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		err := errors.New(http.StatusBadRequest, "Invalid request")
		Error(c, err)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var res Response
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
		assert.False(t, res.Success)
		assert.Equal(t, "Invalid request", res.Message)
		assert.NotNil(t, res.Error)
		assert.Equal(t, http.StatusBadRequest, res.Error.Code)
		assert.Equal(t, "Invalid request", res.Error.Message)
	})

	t.Run("CustomError with wrap", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		underlyingErr := fmt.Errorf("db connection failed")
		err := errors.Wrap(underlyingErr, http.StatusInternalServerError, "Database Error")
		Error(c, err)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var res Response
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
		assert.False(t, res.Success)
		assert.Equal(t, "Database Error", res.Message)
		assert.NotNil(t, res.Error)
		assert.Equal(t, http.StatusInternalServerError, res.Error.Code)
		assert.Equal(t, "Database Error: db connection failed", res.Error.Message)
	})

	t.Run("Generic error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		err := fmt.Errorf("unexpected panic")
		Error(c, err)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var res Response
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &res))
		assert.False(t, res.Success)
		assert.Equal(t, "Internal Server Error", res.Message)
		assert.NotNil(t, res.Error)
		assert.Equal(t, http.StatusInternalServerError, res.Error.Code)
		assert.Equal(t, "unexpected panic", res.Error.Message)
	})
}
