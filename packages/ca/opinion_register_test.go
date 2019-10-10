package ca

import (
	"fmt"
	"testing"
)

func TestOpinionRegister_GetOpinion(t *testing.T) {
	opinionRegister := NewOpinionRegister()

	x := opinionRegister.CreateOpinion("ABC")

	fmt.Println(x.Exists())

	fmt.Println(opinionRegister.GetOpinion("ABC").Exists())
}
