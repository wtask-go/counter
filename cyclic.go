package counter

import (
	"fmt"
	"sync"
)

const (
	// MaxInt - holds maximum int value for your platform
	MaxInt = int(^uint(0) >> 1)
)

// CyclicIncrementor - step-by-step counter with limitation of its maximum value.
// After maximum is reached counter will reset into zero.
// You should use NewCyclicIncrementor() to create counter, but also can create counter like this:
//	c := &counter.CyclicIncrementor{}
// But in that case, counter is not operational until its maximum value will be set:
//	err := c.SetMaxValue(max)
// Also note, if counter is only declared as pointer:
//	var c *CyclicIncrementor
// it is not really initialized and it cannot be used at this point.
type CyclicIncrementor struct {
	mx    sync.RWMutex // for value and max
	value int
	max   int
}

// GetValue - return counter value
func (c *CyclicIncrementor) GetValue() int {
	c.mx.RLock()
	defer c.mx.RUnlock()
	return c.value
}

// Inc - increment by 1 current value of counter. When value is reached max, counter will reset into zero.
func (c *CyclicIncrementor) Inc() {
	c.mx.Lock()
	if c.value < c.max {
		c.value++
	} else {
		c.value = 0
	}
	c.mx.Unlock()
}

// SetMaxValue - change max allowed value for counter.
// Only positive integers allowed to set max value.
func (c *CyclicIncrementor) SetMaxValue(max int) error {
	if max < 0 {
		return fmt.Errorf("counter.CyclicIncrementor: invalid max value (%d)", max)
	}
	c.mx.Lock()
	if c.value > max {
		c.value = 0
	}
	c.max = max
	c.mx.Unlock()
	return nil
}

// NewCyclicIncrementor - return new cyclic counter with preassigned maximum value equals to MaxInt.
func NewCyclicIncrementor() (*CyclicIncrementor, error) {
	c := &CyclicIncrementor{}
	if err := c.SetMaxValue(MaxInt); err != nil {
		return nil, err
	}
	return c, nil
}
