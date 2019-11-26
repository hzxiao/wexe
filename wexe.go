package main


func main() {
	w, err := NewWatcher("./test.txt", NewExecutor("ls"))
	if err != nil {
		panic(err)
	}

	w.Watch()
}
