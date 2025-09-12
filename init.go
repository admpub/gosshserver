package gosshserver

var privatePEM []byte

func initPrivatePEM(keyPath string) error {
	var err error
	privatePEM, err = getPrivatePEM(keyPath)
	return err
}
