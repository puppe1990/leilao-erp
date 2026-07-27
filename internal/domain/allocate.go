package domain

func AllocateUnitCosts(totalCents int64, qty int) []int64 {
	if qty <= 0 {
		return nil
	}
	base := totalCents / int64(qty)
	rem := int(totalCents % int64(qty))
	out := make([]int64, qty)
	for i := 0; i < qty; i++ {
		out[i] = base
		if i < rem {
			out[i]++
		}
	}
	return out
}
