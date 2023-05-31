package common

const debug = true

func DEffect(f func()) {
	if debug {
		f()
	}
}
