package entity

type Stats struct {
	HP, Atk, Def, SpA, SpD, Spe int
}

type TypeSlots struct {
	Primary   int
	Secondary *int
}

type AbilitySlots struct {
	Slot1  int
	Slot2  *int
	Hidden *int
}
