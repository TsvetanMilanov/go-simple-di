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

		Convey("getTags", func() {
			getStructField := func(v interface{}) reflect.StructField {
				return reflect.TypeOf(v).Field(0)
			}

			Convey("Should parse tags correctly when", func() {
				type testCase struct {
					input    interface{}
					expected diTags
				}

				testCases := map[string]testCase{
					"parsing only name.": {
						input: struct {
							F int `di:"name=test"`
						}{},
						expected: diTags{name: "test"},
					},
					"parsing only new with value true.": {
						input: struct {
							F int `di:"new=true"`
						}{},
						expected: diTags{new: true},
					},
					"parsing only new with value false.": {
						input: struct {
							F int `di:"new=false"`
						}{},
						expected: diTags{new: false},
					},
					"parsing only new with invalid value.": {
						input: struct {
							F int `di:"new=false"`
						}{},
						expected: diTags{new: false},
					},
					"parsing all possible values.": {
						input: struct {
							F int `di:"new=true,name=test"`
						}{},
						expected: diTags{new: true, name: "test"},
					},
				}

				for testName, tc := range testCases {
					Convey(testName, func() {
						f := getStructField(tc.input)
						res, err := getTags(f)

						So(err, ShouldBeNil)
						So(*res, ShouldResemble, tc.expected)
					})
				}
			})
			Convey("Should return nil result and nil error when there are no tags.", func() {
				f := getStructField(struct{ F int }{})
				res, err := getTags(f)

				So(err, ShouldBeNil)
				So(res, ShouldBeNil)
			})
			Convey("Should return error when", func() {
				type testCase struct {
					input              interface{}
					expectedErrorValue string
				}

				testCases := map[string]testCase{
					"the tag contains only key.": {
						input: struct {
							F int `di:"key"`
						}{},
						expectedErrorValue: "key",
					},
					"the tag contains only value.": {
						input: struct {
							F int `di:"=value"`
						}{},
						expectedErrorValue: "=value",
					},
					"the tag does not contain value.": {
						input: struct {
							F int `di:"key="`
						}{},
						expectedErrorValue: "key=",
					},
					"the tag does not contain valid key.": {
						input: struct {
							F int `di:"key=value"`
						}{},
						expectedErrorValue: "key=value",
					},
				}

				for testName, tc := range testCases {
					Convey(testName, func() {
						f := getStructField(tc.input)
						res, err := getTags(f)

						So(err, ShouldBeError, getInvalidTagErr(tc.expectedErrorValue))
						So(res, ShouldBeNil)
					})
				}
			})
		})
	})
}
