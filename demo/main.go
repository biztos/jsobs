// demo/main.go

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/biztos/jsobs"
)

// Nb: you can't round-trip nil pointers, but that's a JSON problem not a
// JSOBS problem!
type Thing struct {
	Name string
	Age  float64
}

func (t *Thing) String() string {
	return fmt.Sprintf("{%s age %0.2f}", t.Name, t.Age)
}

func main() {

	client, err := jsobs.NewPgClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Shutdown(1) // in case other shutdown not reached.
	t0 := &Thing{
		Name: "Papa Thing",
		Age:  42,
	}
	t1 := &Thing{
		Name: "Baby Thing",
		Age:  1.23,
	}

	exp := time.Now().Add(time.Second * 2) // to auto-clean at Shutdown

	fmt.Println("storing", t0)
	if err := client.SaveExpiry("/demo/t0.json", t0, exp); err != nil {
		log.Fatal(err)
	}
	fmt.Println("storing", t1)
	if err := client.SaveExpiry("/demo/t1.json", t1, exp); err != nil {
		log.Fatal(err)
	}

	fmt.Println("retrieving")
	r0 := &Thing{}
	if err := client.Load("/demo/t0.json", r0); err != nil {
		log.Fatal(err)
	}
	fmt.Println(r0)
	r1 := &Thing{}
	if err := client.Load("/demo/t1.json", r1); err != nil {
		log.Fatal(err)
	}
	fmt.Println(r1)

	fmt.Println("waiting for expiry")
	time.Sleep(time.Second * 2)

	fmt.Println("shutting down")
	client.Shutdown(0)
}
