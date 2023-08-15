package commands

import "testing"

func TestNewDraftCar(t *testing.T) {
	t.Run("NewDraftCarName", func(t *testing.T) {
		tests := []struct {
			UserID   int64
			Input    string
			Expected string
		}{
			{UserID: 0, Input: "BMW", Expected: "BMW"},
			{UserID: 1, Input: "Lexus", Expected: "Lexus"},
			{UserID: 2, Input: "", Expected: ""},
		}
		c := NewCarCommand(nil)
		for _, test := range tests {
			c.setDraftCarNew(test.UserID)
			c.setDraftCarName(test.UserID, test.Input)
			actual := c.draftCars[test.UserID].Name
			if actual != test.Expected {
				t.Errorf("actual [%s], [%+v]\n", actual, test)
			}
		}
	})

	t.Run("NewDraftCarYear", func(t *testing.T) {
		tests := []struct {
			UserID   int64
			Input    string
			Expected int64
			Error    bool
		}{
			{UserID: 0, Input: "2023", Expected: 2023, Error: false},
			{UserID: 1, Input: "2o23", Expected: 0, Error: true},
		}
		c := NewCarCommand(nil)
		for _, test := range tests {
			c.setDraftCarNew(test.UserID)
			err := c.setDraftCarYear(test.UserID, test.Input)
			if test.Error && err == nil {
				t.Errorf("missing err [%+v]\n", test)
			}
			actual := c.draftCars[test.UserID].Year
			if actual != test.Expected {
				t.Errorf("actual [%d], [%+v]\n", actual, test)
			}
		}
	})

	t.Run("NewDraftCarPlate", func(t *testing.T) {
		tests := []struct {
			UserID   int64
			Input    string
			Expected string
			IsNil    bool
		}{
			{UserID: 0, Input: "/skip", Expected: "", IsNil: true},
			{UserID: 1, Input: "", Expected: "", IsNil: false},
			{UserID: 2, Input: "FZ", Expected: "FZ", IsNil: false},
		}
		c := NewCarCommand(nil)
		for _, test := range tests {
			c.setDraftCarNew(test.UserID)
			c.setDraftCarPlate(test.UserID, test.Input)
			actual := c.draftCars[test.UserID].Plate
			if test.IsNil && actual != nil {
				t.Errorf("actual [%s], [%+v]\n", *actual, test)
			}
			if actual != nil && *actual != test.Expected {
				t.Errorf("actual [%s], [%+v]\n", *actual, test)
			}
		}
	})
}
