package di

import (
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUtils(t *testing.T) {
	Convey("Utils", t, func() {
		Convey("isFieldExported", func() {
			Convey("Should return true for exported struct with exported field.", func() {
				type Exp struct {
					Ex string
				}

				res := isFieldExported(reflect.TypeOf(Exp{}).Field(0))
				So(res, ShouldBeTrue)
			})
			Convey("Should return true for unexported struct with exported field.", func() {
				type exp struct {
					Ex string
				}

				res := isFieldExported(reflect.TypeOf(exp{}).Field(0))
				So(res, ShouldBeTrue)
			})
			Convey("Should return false for Exported struct with unexported field.", func() {
				type Exp struct {
					u string
				}

				res := isFieldExported(reflect.TypeOf(Exp{}).Field(0))
				So(res, ShouldBeFalse)
			})
			Convey("Should return false for unexported struct with unexported field.", func() {
				type un struct {
					u string
				}

				res := isFieldExported(reflect.TypeOf(un{}).Field(0))
				So(res, ShouldBeFalse)
			})
		})
	})
}
