package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"text/template"
)

const N = 5

type Pokemon_info struct {
	Id      int
	Name    string
	Sprites struct {
		Front_default string
	}
	Types []struct {
		Type struct {
			Name string
		}
	}
	Weight float64
}

func Getting_poke(poke_id int) *Pokemon_info {
	response, err := http.Get("https://pokeapi.co/api/v2/pokemon/" + strconv.Itoa(poke_id))
	if err != nil {
		log.Println(err)
		return nil
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return nil
	}

	poke := &Pokemon_info{}

	err = json.Unmarshal(body, poke)
	if err != nil {
		log.Println(err)
		return nil
	}
	return poke
}

func PokeHandler(w http.ResponseWriter, r *http.Request) {
	num_int, err := strconv.Atoi(r.URL.Path[len("/poke/"):])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	poke := Getting_poke(num_int)

	t, err := template.ParseFiles("poke.html")
	if err != nil {
		log.Println(err)
		return
	}
	err = t.Execute(w, poke)
	if err != nil {
		log.Println(err)
		return
	}

}

func Fetch_info(i int, pokes []*Pokemon_info, wg *sync.WaitGroup) {
	poke := Getting_poke(i + 1)
	pokes[i] = poke
	wg.Done()
}

func PokeAllHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("all.html")
	if err != nil {
		log.Println(err)
		return
	}

	pokes := make([]*Pokemon_info, N)
	wg := sync.WaitGroup{}
	for i := 0; i < N; i++ {
		wg.Add(1)
		go Fetch_info(i, pokes, &wg)
	}

	wg.Wait()
	if err = t.Execute(w, pokes); err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/poke/", PokeHandler)
	http.HandleFunc("/poke/all", PokeAllHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}