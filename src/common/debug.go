package common

var debug = false

func init() {
	debug = MustS2I(MustGetEnv("WDM_DEBUG")) != 0
}

func DEffect(f func()) {
	if debug {
		f()
	}
}
