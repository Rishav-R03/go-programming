package main

import (
	"fmt"
	"value_vs_reference/exercises"
)

func primaryExample() {
	a := 10
	b := a
	b = 20
	fmt.Printf("Primitives: a=%d, b=%d\n", a, b)

	type Data struct {
		ID int
	}
	localStruct := Data{ID: 1}
	copyStruct := localStruct

	copyStruct.ID = 9

	fmt.Printf("The local struct ID: %d\n", localStruct.ID)
	fmt.Printf("The copy struct ID : %d\n", copyStruct.ID)
}

type User struct {
	UserID int
	Age    int
}

func referenceExample(user *User, age int) *User {
	user.Age = age
	return user
}

func referenceExample2() {
	//using make() allocates underlying ds in heap
	//name is onto stack and holds only the pointer
	m1 := make(map[string]int)
	m1["key"] = 100
	m2 := m1
	m2["key"] = 101

	fmt.Printf("Before update %d\n", m1["key"])
	fmt.Printf("After update %d\n", m2["key"])
}

func escapeAnalysis() *int {
	x := 50 // stays on stack because it's address is not take or returned
	fmt.Printf("Stack var add: %p\n", &x)
	y := 100
	return &y // escapes to heap
}
func main() {
	primaryExample()
	User1 := User{UserID: 1, Age: 23}
	updatedUser := referenceExample(&User1, 22)
	fmt.Printf("Updated user age: %d", updatedUser.Age)
	referenceExample2()
	ptr := escapeAnalysis()
	fmt.Printf("The pointer %p points to %d\n", ptr, *ptr)

	map1 := make(map[string]int)
	map1["k1"] = 2
	map1["k2"] = 3

	exercises.ExerciseOneGhostUpdate(map1, "k2", 4)
	exercises.ExerciseOneGhostUpdate(map1, "k3", 4)

	configInstance := exercises.Config{ID: 2}
	updatedInstance := exercises.ExerciseTwoSnapShot(&configInstance, 4)
	fmt.Printf("Original instance: %d\n", configInstance.ID)
	fmt.Printf("Updated instance: %d\n", updatedInstance.ID)

	stackVar := exercises.StackVar()
	fmt.Printf("%d\n", stackVar)
	heapVar := exercises.HeapVar()
	fmt.Printf("%d\n", *heapVar)

	exercises.ExerciseFour()

	bool1, bool2 := exercises.ExerciseFive()
	if bool1 {
		fmt.Println("True")
	} else {
		fmt.Println("False")
	}

	if bool2 {
		fmt.Println("True")
	} else {
		fmt.Println("False")
	}

}
