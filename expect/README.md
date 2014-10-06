# expect

Custom assertions for go tests.

## Example

	func TestRun(t *testing.T) {
		expect := expect.New(t)

		expect(nil).ToBeNil()
		expect("").ToNotBeNil()

		expect("test").ToHavePrefix("te")
		expect("test").ToHaveSuffix("st")
		expect("test").ToContain("es")
		expect("a").ToNotBeNil()
		expect("a").ToEqual("a")
		expect("a").ToNotEqual("b")

		expect(int64(1)).ToEqual(1)
		expect(1).ToEqual(1)

		expect("test").ToContain("es")

		arr := []string{"a", "b", "c"}
		expect(arr).ToHaveLength(3)
		expect(arr).ToContain("a")
	}
