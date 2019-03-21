package counter

import (
	"sync"
	"testing"
)

func TestUninitialized(t *testing.T) {
	mustPanic := func(method string) {
		if r := recover(); r == nil {
			t.Errorf("Uninitialized counter (nil) did not panic for %s", method)
		}
	}
	withoutPanic := func(method string) {
		if r := recover(); r != nil {
			t.Errorf("Uninitialized counter (nil) did panic for %s", method)
		}
	}
	cases := []struct {
		testMethod func(c *CyclicIncrementor)
	}{
		{func(c *CyclicIncrementor) { defer mustPanic("GetValue()"); c.GetValue() }},
		{func(c *CyclicIncrementor) { defer mustPanic("Inc()"); c.Inc() }},
		{func(c *CyclicIncrementor) {
			defer withoutPanic("SetMaxValue(-1)")
			if err := c.SetMaxValue(-1); err == nil {
				t.Errorf("SetMaxValue(-1) must return error, but it is not")
			}
		}},
		{func(c *CyclicIncrementor) { defer mustPanic("SetMaxValue(0)"); c.SetMaxValue(0) }},
		{func(c *CyclicIncrementor) { defer mustPanic("SetMaxValue(1)"); c.SetMaxValue(1) }},
	}
	for _, c := range cases {
		c.testMethod(nil)
	}
}

func TestCyclicIncrementor(t *testing.T) {
	// testing in normal flow, without concurrency
	c, err := NewCyclicIncrementor()
	if err != nil {
		t.Errorf("Unexpected initial error: %s", err.Error())
	}

	if c.max != MaxInt {
		t.Errorf("Unexpected initial maximum value (%d)", c.max)
	}

	if c.value != 0 {
		t.Errorf("Unexpected initial current value (%d)", c.value)
	}

	up := 5
	for i := 0; i < up; i++ {
		c.Inc()
	}
	if c.GetValue() != up {
		t.Errorf("Unexpected counter value (%d) after sequential incrementing (%d)", c.GetValue(), up)
	}

	err = c.SetMaxValue(10)
	if err != nil {
		t.Errorf("Unexpected SetMaxValue(10) error: %s", err.Error())
	}
	if c.GetValue() != 5 {
		t.Errorf("Unexpected counter value after maximum changed (%d)", c.GetValue())
	}

	max := 4
	err = c.SetMaxValue(max)
	if err != nil {
		t.Errorf("Unexpected SetMaxValue(%d) error: %s", max, err.Error())
	}
	if c.GetValue() != 0 {
		t.Errorf("Counter value was not reset into zero (%d)", c.GetValue())
	}
	for i := 0; i < max; i++ {
		c.Inc()
	}
	if c.GetValue() != max {
		t.Errorf("Counter value (%d) was not reach allowed maximum (%d)", c.GetValue(), max)
	}

	c.Inc()
	if c.GetValue() != 0 {
		t.Errorf("Counter value (%d) was not reset into zero next after reaching maximum (%d)", c.GetValue(), max)
	}
}

func TestCyclicIncrementorRWConcurrency(t *testing.T) {

	read := func(c *CyclicIncrementor, times int) {
		for i := 0; i < times; i++ {
			// read without sleeping to reach a possible race
			c.GetValue()
		}
	}

	write := func(c *CyclicIncrementor, times int) {
		for i := 0; i < times; i++ {
			// write without sleeping to reach a possible race
			c.Inc()
		}
	}

	c, _ := NewCyclicIncrementor()

	numWriters := 5
	numIncrementsPerWriter := 10
	expectedValue := numWriters * numIncrementsPerWriter
	wg := sync.WaitGroup{}
	wg.Add(numWriters * 2)
	// if race will be detected test will fail
	// with single core the test will not ever fail
	for i := 0; i < numWriters; i++ {
		go func() {
			write(c, numIncrementsPerWriter)
			wg.Done()
		}()
		// also run concurrent reads
		go func() {
			read(c, 20)
			wg.Done()
		}()
	}
	wg.Wait()

	if c.GetValue() != expectedValue {
		t.Errorf("Unexpected counter value (%d)", c.GetValue())
	}
}
