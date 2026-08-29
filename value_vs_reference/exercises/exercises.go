package exercises

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func ExerciseOneGhostUpdate(inputMap map[string]int, key string, val int) {
	if inputMap[key] == 0 {
		fmt.Printf("The key %s is not present in map\n", key)
	}
	inputMap[key] = val
}

type Config struct {
	ID int
}

func ExerciseTwoSnapShot(localConfig *Config, update int) *Config {
	localConfig.ID = update
	return localConfig
}

func StackVar() int {
	max := big.NewInt(100)
	randInt, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0
	}
	return int(randInt.Int64())
}

func HeapVar() *int {
	max := big.NewInt(1000)
	randInt, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil
	}
	i := int(randInt.Int64())
	return &i
}

func ExerciseFour() {
	slice1 := make([]int, 0, 5)
	slice1 = append(slice1, 1)
	slice1 = append(slice1, 2)
	slice1 = append(slice1, 3)
	fmt.Println("The slice 1 ", slice1)
	slice2 := slice1
	slice1 = append(slice1, 4)
	fmt.Println("The slice 1 ", slice1)
	fmt.Println("The slice 2 ", slice2)
}

func ExerciseFive() (bool, bool) {
	ch1 := make(chan int)
	ch2 := ch1
	ch3 := make(chan int)
	return ch1 == ch2, ch1 == ch3
}
