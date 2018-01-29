package models

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestValidateStateFilter(t *testing.T) {
	t.Parallel()
	Convey("Successfully return without any errors", t, func() {
		Convey("when the filter list contains a state of `created`", func() {

			err := ValidateStateFilter([]string{CreatedState})
			So(err, ShouldBeNil)
		})

		Convey("when the filter list contains a state of `submitted`", func() {

			err := ValidateStateFilter([]string{SubmittedState})
			So(err, ShouldBeNil)
		})

		Convey("when the filter list contains a state of `completed`", func() {

			err := ValidateStateFilter([]string{CompletedState})
			So(err, ShouldBeNil)
		})

		Convey("when the filter list contains a state of `edition-confirmed`", func() {

			err := ValidateStateFilter([]string{EditionConfirmedState})
			So(err, ShouldBeNil)
		})

		Convey("when the filter list contains a state of `associated`", func() {

			err := ValidateStateFilter([]string{AssociatedState})
			So(err, ShouldBeNil)
		})

		Convey("when the filter list contains a state of `published`", func() {

			err := ValidateStateFilter([]string{PublishedState})
			So(err, ShouldBeNil)
		})

		Convey("when the filter list contains more than one valid state", func() {

			err := ValidateStateFilter([]string{EditionConfirmedState, PublishedState})
			So(err, ShouldBeNil)
		})
	})

	Convey("Return with errors", t, func() {
		Convey("when the filter list contains an invalid state", func() {

			err := ValidateStateFilter([]string{"foo"})
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("bad request - invalid filter state values: [foo]"))
		})

		Convey("when the filter list contains more than one invalid state", func() {

			err := ValidateStateFilter([]string{"foo", "bar"})
			So(err, ShouldNotBeNil)
			So(err, ShouldResemble, errors.New("bad request - invalid filter state values: [foo bar]"))
		})
	})
}
