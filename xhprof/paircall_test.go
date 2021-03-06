package xhprof

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePairName(t *testing.T) {
	cases := []struct {
		Name   string
		Parent string
		Child  string
	}{
		{
			Name:   "main()",
			Parent: "",
			Child:  "main()",
		},
		{
			Name:   "main()==>foo",
			Parent: "main()",
			Child:  "foo",
		},
	}

	for _, c := range cases {
		parent, child := parsePairName(c.Name)
		assert.Equal(t, c.Parent, parent)
		assert.Equal(t, c.Child, child)
	}
}

func TestFlatten(t *testing.T) {
	expected := &Profile{
		Calls: []*Call{
			&Call{
				Name:              "main()",
				Count:             1,
				WallTime:          1000,
				ExclusiveWallTime: 500,
				CpuTime:           400,
				ExclusiveCpuTime:  200,
				Memory:            1500,
				ExclusiveMemory:   800,
				IoTime:            600,
				ExclusiveIoTime:   300,
			},
			&Call{
				Name:              "foo",
				Count:             2,
				WallTime:          500,
				ExclusiveWallTime: 300,
				CpuTime:           200,
				ExclusiveCpuTime:  100,
				Memory:            700,
				ExclusiveMemory:   400,
				IoTime:            300,
				ExclusiveIoTime:   200,
			},
			&Call{
				Name:              "bar",
				Count:             10,
				WallTime:          200,
				ExclusiveWallTime: 200,
				CpuTime:           100,
				ExclusiveCpuTime:  100,
				Memory:            300,
				ExclusiveMemory:   300,
				IoTime:            100,
				ExclusiveIoTime:   100,
			},
		},
	}

	m := &PairCallMap{
		M: map[string]*PairCall{
			"main()": &PairCall{
				WallTime: 1000,
				Count:    1,
				CpuTime:  400,
				Memory:   1500,
			},
			"main()==>foo": &PairCall{
				WallTime: 500,
				Count:    2,
				CpuTime:  200,
				Memory:   700,
			},
			"foo==>bar": &PairCall{
				WallTime: 200,
				Count:    10,
				CpuTime:  100,
				Memory:   300,
			},
		},
	}

	profile := m.Flatten()
	require.IsType(t, profile, expected)

	profile.SortBy("WallTime")

	assert.Equal(t, float32(1000), profile.Main.WallTime)
	assert.EqualValues(t, expected.Calls, profile.Calls)
}

func TestAvgPairCallMaps(t *testing.T) {
	expected := &PairCallMap{
		M: map[string]*PairCall{
			"main()": &PairCall{
				WallTime: 600,
				Count:    1,
				CpuTime:  300,
				Memory:   700,
			},
			"main()==>foo": &PairCall{
				WallTime: 300,
				Count:    2,
				CpuTime:  170,
				Memory:   500,
			},
			"foo==>bar": &PairCall{
				WallTime: 100,
				Count:    3,
				CpuTime:  50,
				Memory:   100,
			},
		},
	}
	m1 := &PairCallMap{
		M: map[string]*PairCall{
			"main()": &PairCall{
				WallTime: 800,
				Count:    1,
				CpuTime:  400,
				Memory:   1000,
			},
			"main()==>foo": &PairCall{
				WallTime: 600,
				Count:    2,
				CpuTime:  300,
				Memory:   900,
			},
			"foo==>bar": &PairCall{
				WallTime: 300,
				Count:    10,
				CpuTime:  150,
				Memory:   300,
			},
		},
	}
	m2 := &PairCallMap{
		M: map[string]*PairCall{
			"main()": &PairCall{
				WallTime: 300,
				Count:    1,
				CpuTime:  100,
				Memory:   200,
			},
		},
	}
	m3 := &PairCallMap{
		M: map[string]*PairCall{
			"main()": &PairCall{
				WallTime: 700,
				Count:    1,
				CpuTime:  400,
				Memory:   900,
			},
			"main()==>foo": &PairCall{
				WallTime: 300,
				Count:    4,
				CpuTime:  210,
				Memory:   600,
			},
		},
	}

	res := AvgPairCallMaps([]*PairCallMap{m1, m2, m3})
	assert.EqualValues(t, expected, res)
}

func TestComputeNearestFamily(t *testing.T) {
	expected := &NearestFamily{
		Children: &PairCallMap{
			M: map[string]*PairCall{
				"bar": &PairCall{
					WallTime: 200,
					Count:    10,
				},
			},
		},
		Parents: &PairCallMap{
			M: map[string]*PairCall{
				"main()": &PairCall{
					WallTime: 500,
					Count:    2,
				},
			},
		},
		ChildrenCount: 10,
		ParentsCount:  2,
	}

	m := &PairCallMap{
		M: map[string]*PairCall{
			"main()": &PairCall{
				WallTime: 1000,
				Count:    1,
				CpuTime:  400,
				Memory:   1500,
			},
			"main()==>foo": &PairCall{
				WallTime: 500,
				Count:    2,
				CpuTime:  200,
				Memory:   700,
			},
			"foo==>bar": &PairCall{
				WallTime: 200,
				Count:    10,
				CpuTime:  100,
				Memory:   300,
			},
		},
	}

	f := m.ComputeNearestFamily("foo")

	assert.EqualValues(t, expected, f)
}

func TestSubtractPairCallMaps(t *testing.T) {
	expected := &PairCallMap{
		M: map[string]*PairCall{
			"main()": &PairCall{
				WallTime: 500,
				Count:    0,
				CpuTime:  300,
				Memory:   800,
			},
			"main()==>foo": &PairCall{
				WallTime: 600,
				Count:    2,
				CpuTime:  300,
				Memory:   900,
			},
			"foo==>bar": &PairCall{
				WallTime: -300,
				Count:    -10,
				CpuTime:  -150,
				Memory:   -300,
			},
		},
	}
	m1 := &PairCallMap{
		M: map[string]*PairCall{
			"main()": &PairCall{
				WallTime: 800,
				Count:    1,
				CpuTime:  400,
				Memory:   1000,
			},
			"main()==>foo": &PairCall{
				WallTime: 600,
				Count:    2,
				CpuTime:  300,
				Memory:   900,
			},
		},
	}
	m2 := &PairCallMap{
		M: map[string]*PairCall{
			"main()": &PairCall{
				WallTime: 300,
				Count:    1,
				CpuTime:  100,
				Memory:   200,
			},
			"foo==>bar": &PairCall{
				WallTime: 300,
				Count:    10,
				CpuTime:  150,
				Memory:   300,
			},
		},
	}

	diff := m1.Subtract(m2)
	assert.EqualValues(t, expected, diff)
}
