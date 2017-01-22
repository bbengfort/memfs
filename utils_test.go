package memfs_test

import (
	. "github.com/bbengfort/memfs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {

	Describe("string helpers", func() {

		It("should regularize a string", func() {

			var regTests = []struct {
				testcase string // input to func
				expected string // expected result
			}{
				{"tomato", "tomato"},
				{"Apple", "apple"},
				{"ORANGE", "orange"},
				{" pear", "pear"},
				{"banana ", "banana"},
				{" blueberry ", "blueberry"},
				{"    canteloupe   ", "canteloupe"},
				{"  PePPer   ", "pepper"},
				{" going to Giant   ", "going to giant"},
				{"\n lines are bad \n", "lines are bad"},
				{"\t\t tabs are bad \t\t", "tabs are bad"},
				{"\n\r\t    all whitespace \n\r\t\n\n", "all whitespace"},
			}

			for _, tt := range regTests {
				actual := Regularize(tt.testcase)
				Ω(actual).Should(Equal(tt.expected))
			}

		})

		It("should stride over a string", func() {

			var strideTests = []struct {
				s string   // string to split
				n int      // amount to stride by
				e []string // expected results
			}{
				{"the eagle flies at midnight", 6, []string{"the ea", "gle fl", "ies at", " midni", "ght"}},
				{"the eagle flies at midnight", 8, []string{"the eagl", "e flies ", "at midni", "ght"}},
				{"the eagle flies at midnight", 9, []string{"the eagle", " flies at", " midnight"}},
				{"the eagle flies at midnight", 10, []string{"the eagle ", "flies at m", "idnight"}},
				{"the eagle flies at midnight", 14, []string{"the eagle flie", "s at midnight"}},
			}

			for _, tt := range strideTests {
				Ω(Stride(tt.s, tt.n)).Should(Equal(tt.e))
			}

		})

		It("should stride fixed length substrings over a string", func() {

			var strideTests = []struct {
				s string   // string to split
				n int      // amount to stride by
				e []string // expected results
			}{
				{"the eagle flies at midnight", 6, []string{"the ea", "gle fl", "ies at", " midni"}},
				{"the eagle flies at midnight", 8, []string{"the eagl", "e flies ", "at midni"}},
				{"the eagle flies at midnight", 9, []string{"the eagle", " flies at", " midnight"}},
				{"the eagle flies at midnight", 10, []string{"the eagle ", "flies at m"}},
				{"the eagle flies at midnight", 14, []string{"the eagle flie"}},
				{"the eagle flies at midnight", 27, []string{"the eagle flies at midnight"}},
				{"the eagle flies at midnight", 28, []string{}},
			}

			for _, tt := range strideTests {
				Ω(StrideFixed(tt.s, tt.n)).Should(Equal(tt.e))
			}

		})
	})

	Describe("collection helpers", func() {

		It("should determine of a string is contained in a list", func() {
			items := []string{"poblano", "serrano", "cayenne", "habanero"}
			Ω(ListContains("cayenne", items)).Should(BeTrue())
			Ω(ListContains("vanilla", items)).Should(BeFalse())
		})

	})

	Describe("numeric helpers", func() {

		It("should determine the maximal value of uints", func() {
			testCases := []struct {
				vals []uint64
				max  uint64
			}{
				{[]uint64{}, 0},
				{[]uint64{319922, 319922, 319922}, 319922},
				{[]uint64{434930, 434931}, 434931},
				{[]uint64{100, 42, 810, 321, 1, 0, 3193, 2}, 3193},
			}

			for _, tt := range testCases {
				Ω(MaxUInt64(tt.vals...)).Should(Equal(tt.max))
			}

		})

	})

})
