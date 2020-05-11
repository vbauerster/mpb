package internal

func CalcWidthForBarFiller(reqWidth, available int) int {
	if reqWidth <= 0 || reqWidth >= available {
		return available
	}
	return reqWidth
}
