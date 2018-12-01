package types_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/nickhstr/goweb/types"
	c "github.com/smartystreets/goconvey/convey"
)

func TestErrorResponse(t *testing.T) {
	c.Convey("Given an error", t, func() {
		errMsg := "Something has gone horribly wrong"
		newErr := errors.New(errMsg)

		c.Convey("When it needs to be marshaled", func() {
			c.Convey("It should be valid JSON", func() {
				errResp := types.NewErrorResponse(newErr)
				actual := string(errResp)
				expected := fmt.Sprintf(`{"error":"%s"}`, errMsg)

				c.So(actual, c.ShouldEqual, expected)
			})
		})
	})
}
