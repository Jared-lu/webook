package main

func main() {
	server := initApp()
	err := server.Run(":8080")
	if err != nil {
		return
	}
}
