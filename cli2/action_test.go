package cli2

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func testCreateAction(path string, r Runner) *action {
	return &action{path: path, runner: r, params: map[string]*option{}}
}

type ActionWithoutType struct {
	Field bool `cli:"no='type set'"`
}

func (a *ActionWithoutType) Run() error {
	return nil
}

func TestActionWithoutType(t *testing.T) {
	Convey("Given an action with an unknown type", t, func() {
		Convey("When the reflect method is called on it", func() {
			a := testCreateAction("some/path", &ActionWithoutType{})
			e := a.reflect()
			Convey("Then there is an error", func() {
				So(e, ShouldNotBeNil)
				So(e.Error(), ShouldEqual, `ActionWithoutType: tag for field "Field" has no type set`)
			})
		})
	})
}

type ActionWithWrongType struct {
	Field bool `cli:"type=unknown"`
}

func (a *ActionWithWrongType) Run() error {
	return nil
}

func TestActionWithWrongType(t *testing.T) {
	Convey("Given an action with an unknown type", t, func() {
		Convey("When the reflect method is called on it", func() {
			a := testCreateAction("some/path", &ActionWithWrongType{})
			e := a.reflect()
			Convey("Then there is an error", func() {
				So(e, ShouldNotBeNil)
				So(e.Error(), ShouldEqual, `ActionWithWrongType: tag for field "Field" has unknown type "unknown"`)
			})
		})
	})
}

type ActionWithFlag struct {
	Flag1 bool `cli:"type=opt short=f long=flag"`
}

func (a *ActionWithFlag) Run() error {
	return nil
}

func TestActionWithFlag(t *testing.T) {
	Convey("Given an action with a valid flag", t, func() {
		Convey("When the reflect method is called on it", func() {
			baseAction := &ActionWithFlag{}
			a := testCreateAction("some/path2", baseAction)
			e := a.reflect()
			Convey("Then there is no error", func() {
				So(e, ShouldBeNil)
			})
			Convey("And there is a flag defined for the action", func() {
				So(len(a.opts), ShouldEqual, 1)
				So(a.opts[0].short, ShouldEqual, "f")
				So(a.opts[0].long, ShouldEqual, "flag")
				So(a.opts[0].isFlag, ShouldBeTrue)
			})
			Convey("When empty arguments are parsed", func() {
				a.parseArgs([]string{})
				Convey("Then the flag is not set", func() {
					So(baseAction.Flag1, ShouldBeFalse)
				})
			})
			Convey("When arguments are parsed with Flag1 set (using short option)", func() {
				a.opts[0].value = "" // reset value.
				a.parseArgs([]string{"-f"})
				Convey("Then the flag is set", func() {
					So(baseAction.Flag1, ShouldBeTrue)
				})
			})
			Convey("When arguments are parsed with Flag1 set (using long option)", func() {
				a.opts[0].value = "" // reset value.
				a.parseArgs([]string{"--flag"})
				Convey("Then the flag is set", func() {
					So(baseAction.Flag1, ShouldBeTrue)
				})
			})
		})
	})
}

type ActionWithFlag_Error1 struct {
	Flag1 bool `cli:"type=opt"`
}

func (a *ActionWithFlag_Error1) Run() error {
	return nil
}

func TestActionWithFlagError1(t *testing.T) {
	Convey("Given an action with a valid flag", t, func() {
		Convey("When the reflect method is called on it", func() {
			a := testCreateAction("some/path", &ActionWithFlag_Error1{})
			e := a.reflect()
			Convey("Then there is an error", func() {
				So(e, ShouldNotBeNil)
				So(e.Error(), ShouldEqual, `ActionWithFlag_Error1: option "Flag1" has neither long nor short accessor set`)
			})
			Convey("And there is no flag defined for the action", func() {
				So(len(a.opts), ShouldEqual, 0)
			})
		})
	})
}

type ActionWithOption struct {
	Option int `cli:"type=opt short=o long=option"`
}

func (a *ActionWithOption) Run() error {
	return nil
}

func TestActionWithOption(t *testing.T) {
	Convey("Given an action with a valid option", t, func() {
		Convey("When the reflect method is called on it", func() {
			baseAction := &ActionWithOption{}
			a := testCreateAction("some/path", baseAction)
			e := a.reflect()
			Convey("Then there is no error", func() {
				So(e, ShouldBeNil)
			})
			Convey("And there is an option defined for the action", func() {
				So(len(a.opts), ShouldEqual, 1)
			})

			Convey("When empty arguments are parsed", func() {
				a.parseArgs([]string{})
				Convey("Then the option is not set", func() {
					So(baseAction.Option, ShouldEqual, 0)
				})
			})
			Convey("When arguments are parsed with Option set (using short option)", func() {
				a.parseArgs([]string{"-o", "1"})
				Convey("Then the option is set to 1", func() {
					So(baseAction.Option, ShouldEqual, 1)
				})
			})
			Convey("When arguments are parsed with Option set (using long option)", func() {
				a.parseArgs([]string{"--option", "2"})
				Convey("Then the option is set to 2", func() {
					So(baseAction.Option, ShouldEqual, 2)
				})
			})
		})
	})
}

type ActionWithRequiredOption struct {
	Option int `cli:"type=opt short=o long=option required=true"`
}

func (a *ActionWithRequiredOption) Run() error {
	return nil
}

func TestActionWithRequiredOption(t *testing.T) {
	Convey("Given an action with a required option", t, func() {
		Convey("When the reflect method is called on it", func() {
			baseAction := &ActionWithRequiredOption{}
			a := testCreateAction("some/path", baseAction)
			e := a.reflect()
			Convey("Then there is no error", func() {
				So(e, ShouldBeNil)
			})
			Convey("And there is an option defined for the action", func() {
				So(len(a.opts), ShouldEqual, 1)
			})
			Convey("And the option defined has the required flag set", func() {
				So(a.opts[0].required, ShouldBeTrue)
			})

			Convey("When empty arguments are parsed", func() {
				e := a.parseArgs([]string{})
				Convey("Then there is an error", func() {
					So(e, ShouldNotBeNil)
					So(e.Error(), ShouldEqual, `option "Option" is required but not set`)
				})
			})

			Convey("When arguments are parsed with Option set (using short option)", func() {
				a.parseArgs([]string{"-o", "1"})
				Convey("Then the option is set to 1", func() {
					So(baseAction.Option, ShouldEqual, 1)
				})
			})
		})
	})
}

type ActionWithOptionWithDefault struct {
	Option1 int    `cli:"type=opt short=o long=optionA default=1"`
	Option2 string `cli:"type=opt short=p long=optionB default=x"`
}

func (a *ActionWithOptionWithDefault) Run() error {
	return nil
}

func TestActionWithOptionWithDefaults(t *testing.T) {
	Convey("Given an action with a valid option with defaults", t, func() {
		Convey("When the reflect method is called on it", func() {
			baseAction := &ActionWithOptionWithDefault{}
			a := testCreateAction("some/path", baseAction)
			e := a.reflect()
			Convey("Then there is no error", func() {
				t.Log(e)
				So(e, ShouldBeNil)
			})
			Convey("And there is an option defined for the action", func() {
				So(len(a.opts), ShouldEqual, 2)
			})
			Convey("And there is are defaults defined for the options", func() {
				So(a.opts[0].value, ShouldEqual, "1")
				So(a.opts[1].value, ShouldEqual, "x")
			})

			Convey("When empty arguments are parsed", func() {
				e := a.parseArgs([]string{})

				Convey("Then there is no error", func() {
					So(e, ShouldBeNil)
				})

				Convey("Then the option is set to 1", func() {
					So(baseAction.Option1, ShouldEqual, 1)
					So(baseAction.Option2, ShouldEqual, "x")
				})
			})

			Convey("When arguments are parsed with options set (using short option)", func() {
				e := a.parseArgs([]string{"-o", "42", "-p", "foobar"})

				Convey("Then there is no error", func() {
					So(e, ShouldBeNil)
				})

				Convey("Then the options are set", func() {
					So(baseAction.Option1, ShouldEqual, 42)
					So(baseAction.Option2, ShouldEqual, "foobar")
				})
			})
		})
	})
}

type ActionWithOption_Error1 struct {
	Option int `cli:"type=opt"`
}

func (a *ActionWithOption_Error1) Run() error {
	return nil
}

func TestActionWithOptionError1(t *testing.T) {
	Convey("Given an action with a valid option", t, func() {
		Convey("When the reflect method is called on it", func() {
			a := testCreateAction("some/path", &ActionWithOption_Error1{})
			e := a.reflect()
			Convey("Then there is an error", func() {
				So(e, ShouldNotBeNil)
				So(e.Error(), ShouldEqual, `ActionWithOption_Error1: option "Option" has neither long nor short accessor set`)
			})
			Convey("And there is no option defined for the action", func() {
				So(len(a.opts), ShouldEqual, 0)
			})
		})
	})
}

type ActionWithOption_Error2 struct {
	Option int `cli:"type=opt short=o required=invalid"`
}

func (a *ActionWithOption_Error2) Run() error {
	return nil
}

func TestActionWithOptionError2(t *testing.T) {
	Convey("Given an action with a invalid value for the required tag", t, func() {
		Convey("When the reflect method is called on it", func() {
			a := testCreateAction("some/path", &ActionWithOption_Error2{})
			e := a.reflect()
			Convey("Then there is an error", func() {
				So(e, ShouldNotBeNil)
				So(e.Error(), ShouldEqual, `ActionWithOption_Error2: wrong value for "required" tag: "invalid"`)
			})
			Convey("And there is no option defined for the action", func() {
				So(len(a.opts), ShouldEqual, 0)
			})
		})
	})
}

type ActionWithOption_Error3 struct {
	Option int `cli:"type=opt short=o default=noInt"`
}

func (a *ActionWithOption_Error3) Run() error {
	return nil
}

func TestActionWithOptionError3(t *testing.T) {
	Convey("Given an action with a invalid value for the default tag (not an int)", t, func() {
		Convey("When the reflect method is called on it", func() {
			a := testCreateAction("some/path", &ActionWithOption_Error3{})
			e := a.reflect()
			Convey("Then there is an error", func() {
				So(e, ShouldNotBeNil)
				So(e.Error(), ShouldEqual, `ActionWithOption_Error3: wrong value for "default" tag: strconv.ParseInt: parsing "noInt": invalid syntax`)
			})
			Convey("And there is no option defined for the action", func() {
				So(len(a.opts), ShouldEqual, 0)
			})
		})
	})
}

type ActionWithArguments struct {
	Argument  int    `cli:"type=arg"`
	StringArg string `cli:"type=arg"`
}

func (a *ActionWithArguments) Run() error {
	return nil
}

func TestActionWithArguments(t *testing.T) {
	Convey("Given an action with an argument", t, func() {
		Convey("When the reflect method is called on it", func() {
			baseAction := &ActionWithArguments{}
			a := testCreateAction("some/path", baseAction)
			e := a.reflect()

			Convey("Then there is no error", func() {
				So(e, ShouldBeNil)
			})
			Convey("And there is an argument defined for the action", func() {
				So(len(a.args), ShouldEqual, 2)
			})

			Convey("When empty arguments are parsed", func() {
				e := a.parseArgs([]string{})

				Convey("Then there is no error", func() {
					So(e, ShouldBeNil)
				})

				Convey("Then the argument is set to 0", func() {
					So(baseAction.Argument, ShouldEqual, 0)
				})
			})

			Convey("When arguments are given", func() {
				e := a.parseArgs([]string{"1", "test"})

				Convey("Then there is no error", func() {
					So(e, ShouldBeNil)
				})

				Convey("Then the argument is set to 1", func() {
					So(baseAction.Argument, ShouldEqual, 1)
					So(baseAction.StringArg, ShouldEqual, "test")
				})
			})

			Convey("When arguments with wrong type are given", func() {
				e := a.parseArgs([]string{"x"})

				Convey("Then there is an error", func() {
					So(e, ShouldNotBeNil)
					So(e.Error(), ShouldEqual, `argument "Argument" at index "0" has wrong type`)
				})

				Convey("When too many arguments are given", func() {
					e := a.parseArgs([]string{"1", "2", "3"})

					Convey("Then there is an error", func() {
						So(e, ShouldNotBeNil)
						So(e.Error(), ShouldEqual, "too many arguments given")
					})
				})
			})
		})
	})
}

type ActionWithRequiredArgument struct {
	Argument int `cli:"type=arg required=true"`
}

func (a *ActionWithRequiredArgument) Run() error {
	return nil
}

func TestActionWithRequiredArgument(t *testing.T) {
	Convey("Given an action with an argument", t, func() {
		Convey("When the reflect method is called on it", func() {
			baseAction := &ActionWithRequiredArgument{}
			a := testCreateAction("some/path", baseAction)
			e := a.reflect()

			Convey("Then there is no error", func() {
				So(e, ShouldBeNil)
			})
			Convey("And there is an argument defined for the action", func() {
				So(len(a.args), ShouldEqual, 1)
			})

			Convey("When empty arguments are parsed", func() {
				e := a.parseArgs([]string{})

				Convey("Then there is an error", func() {
					So(e, ShouldNotBeNil)
					So(e.Error(), ShouldEqual, "required argument not set")
				})
			})

			Convey("When argument is given", func() {
				e := a.parseArgs([]string{"1"})

				Convey("Then there is no error", func() {
					So(e, ShouldBeNil)
				})

				Convey("Then the argument is set to 1", func() {
					So(baseAction.Argument, ShouldEqual, 1)
				})
			})
		})
	})
}

type ActionWithMultipleArguments struct {
	Argument1 int    `cli:"type=arg"`
	Argument2 string `cli:"type=arg"`
}

func (a *ActionWithMultipleArguments) Run() error {
	return nil
}

func TestActionWithMultipleArguments(t *testing.T) {
	Convey("Given an action with an argument", t, func() {
		Convey("When the reflect method is called on it", func() {
			baseAction := &ActionWithMultipleArguments{}
			a := testCreateAction("some/path", baseAction)
			e := a.reflect()

			Convey("Then there is no error", func() {
				So(e, ShouldBeNil)
			})
			Convey("And there is an argument defined for the action", func() {
				So(len(a.args), ShouldEqual, 2)
			})

			Convey("When empty arguments are parsed", func() {
				e := a.parseArgs([]string{})

				Convey("Then there is no error", func() {
					So(e, ShouldBeNil)
				})
			})

			Convey("When only first argument is given", func() {
				e := a.parseArgs([]string{"1"})

				Convey("Then there is no error", func() {
					So(e, ShouldBeNil)
				})

				Convey("Then the argument is set to 1", func() {
					So(baseAction.Argument1, ShouldEqual, 1)
					So(baseAction.Argument2, ShouldEqual, "")
				})
			})

			Convey("When both arguments are given", func() {
				e := a.parseArgs([]string{"42", "foobar"})

				Convey("Then there is no error", func() {
					So(e, ShouldBeNil)
				})

				Convey("Then the argument is set to 1", func() {
					So(baseAction.Argument1, ShouldEqual, 42)
					So(baseAction.Argument2, ShouldEqual, "foobar")
				})
			})
		})
	})
}

type ActionWithVariadicArgument struct {
	Argument1 int      `cli:"type=arg"`
	Argument2 []string `cli:"type=arg"`
}

func (a *ActionWithVariadicArgument) Run() error {
	return nil
}

func TestActionWithVariadicArgument(t *testing.T) {
	Convey("Given an action with an argument", t, func() {
		Convey("When the reflect method is called on it", func() {
			baseAction := &ActionWithVariadicArgument{}
			a := testCreateAction("some/path", baseAction)
			e := a.reflect()

			Convey("Then there is no error", func() {
				So(e, ShouldBeNil)
			})
			Convey("And there is an argument defined for the action", func() {
				So(len(a.args), ShouldEqual, 2)
			})

			Convey("When empty arguments are parsed", func() {
				e := a.parseArgs([]string{})

				Convey("Then there is no error", func() {
					So(e, ShouldBeNil)
				})
			})

			Convey("When only first argument is given", func() {
				e := a.parseArgs([]string{"1"})

				Convey("Then there is no error", func() {
					So(e, ShouldBeNil)
				})

				Convey("Then the argument is set to 1", func() {
					So(baseAction.Argument1, ShouldEqual, 1)
					So(len(baseAction.Argument2), ShouldEqual, 0)
				})
			})
		})

		Convey("When both arguments are given (with single values for variadic argument)", func() {
			baseAction := &ActionWithVariadicArgument{}
			a := testCreateAction("some/path", baseAction)
			_ = a.reflect()
			e := a.parseArgs([]string{"42", "foobar"})

			Convey("Then there is no error", func() {
				So(e, ShouldBeNil)
			})

			Convey("Then the argument is set to 1", func() {
				So(baseAction.Argument1, ShouldEqual, 42)
				So(len(baseAction.Argument2), ShouldEqual, 1)
				So("foobar", ShouldBeIn, baseAction.Argument2)
			})
		})

		Convey("When both arguments are given (with multiple values for variadic argument)", func() {
			baseAction := &ActionWithVariadicArgument{}
			a := testCreateAction("some/path", baseAction)
			_ = a.reflect()
			e := a.parseArgs([]string{"42", "foobar", "foobaz", "fuuboz"})

			Convey("Then there is no error", func() {
				So(e, ShouldBeNil)
			})

			Convey("Then the argument is set to 1", func() {
				So(baseAction.Argument1, ShouldEqual, 42)
				t.Log(baseAction.Argument2)
				So(len(baseAction.Argument2), ShouldEqual, 3)
				So("foobar", ShouldBeIn, baseAction.Argument2)
				So("foobaz", ShouldBeIn, baseAction.Argument2)
				So("fuuboz", ShouldBeIn, baseAction.Argument2)
			})
		})
	})
}

type ActionWithArgument_Error1 struct {
	Argument1 []string `cli:"type=arg"`
	Argument2 int      `cli:"type=arg"`
}

func (a *ActionWithArgument_Error1) Run() error {
	return nil
}

func TestActionWithArgument_Error1(t *testing.T) {
	Convey("Given an action with an argument", t, func() {
		Convey("When the reflect method is called on it", func() {
			baseAction := &ActionWithArgument_Error1{}
			a := testCreateAction("some/path", baseAction)
			e := a.reflect()

			Convey("Then there is an error", func() {
				So(e, ShouldNotBeNil)
				So(e.Error(), ShouldEqual, "ActionWithArgument_Error1: only last argument can be variadic")
			})
		})
	})
}
