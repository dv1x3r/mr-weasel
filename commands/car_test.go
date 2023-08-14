package commands

// import "testing"

// func TestAskNewCar(t *testing.T) {
// 	t.Run("WithPlate", func(t *testing.T) {
// 		c := NewCarCommand(nil)
// 		c.askNewCarStart(nil, Payload{UserID: 1})
// 		c.askNewCarName(nil, Payload{UserID: 1, Command: "ABC"})  // Name
// 		c.askNewCarYear(nil, Payload{UserID: 1, Command: "2023"}) // Year
// 		c.askNewCarPlates(Payload{UserID: 1, Command: "FZ"})      // Plate
// 		if c.draftCars[1].Name != "ABC" {
// 			t.Errorf("expected name [ABC], actual [%s]\n", c.draftCars[1].Name)
// 		}
// 		if c.draftCars[1].Year != 2023 {
// 			t.Errorf("expected year [2023], actual [%d]\n", c.draftCars[1].Year)
// 		}
// 		if *c.draftCars[1].Plate != "FZ" {
// 			t.Errorf("expected car name [FZ], actual [%s]\n", *c.draftCars[1].Plate)
// 		}
// 	})
// 	t.Run("NoPlate", func(t *testing.T) {
// 		c := NewCarCommand(nil)
// 		c.askNewCarStart(nil, Payload{UserID: 1})
// 		c.askNewCarName(nil, Payload{UserID: 1, Command: "ABC"})  // Name
// 		c.askNewCarYear(nil, Payload{UserID: 1, Command: "2023"}) // Year
// 		c.askNewCarPlates(Payload{UserID: 1, Command: "/skip"})   // Plate
// 		if c.draftCars[1].Name != "ABC" {
// 			t.Errorf("expected name [ABC], actual [%s]\n", c.draftCars[1].Name)
// 		}
// 		if c.draftCars[1].Year != 2023 {
// 			t.Errorf("expected year [2023], actual [%d]\n", c.draftCars[1].Year)
// 		}
// 		if c.draftCars[1].Plate != nil {
// 			t.Errorf("expected car name [nil], actual [%s]\n", *c.draftCars[1].Plate)
// 		}
// 	})
// }
