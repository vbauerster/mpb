package decor

// Marquee returns marquee decorator that will scroll text right to left.
//
// `text` is the scrolling message
//
// `ws` controls the showing window size
func Marquee(text string, ws uint, wcc ...WC) Decorator {
	bytes := []byte(text)
	buf := make([]byte, ws)
	var count uint
	f := func(s Statistics) string {
		start := count % uint(len(bytes))
		var i uint = 0
		for ; i < ws && start+i < uint(len(bytes)); i++ {
			buf[i] = bytes[start+i]
		}
		for ; i < ws; i++ {
			buf[i] = ' '
		}
		count++
		return string(buf)
	}
	return Any(f, wcc...)
}
