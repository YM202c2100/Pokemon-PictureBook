package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"text/template"

	"gopkg.in/ini.v1"
)

type Config_list struct {
	N int
}

var cfg Config_list

func init() {
	config, err := ini.Load("config.ini")
	if err != nil {
		log.Fatal(err)
	}
	cfg = Config_list{
		N: config.Section("max_display").Key("n").MustInt(),
	}
}

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

	t, err := template.ParseFiles("html/poke.html")
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

//No.がi+1のポケモンの情報を得る
func Fetch_info(i int, pokes []*Pokemon_info, wg *sync.WaitGroup) {
	poke := Getting_poke(i + 1)
	pokes[i] = poke
	wg.Done()
}

func PokeTableHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("html/table.html")
	if err != nil {
		log.Println(err)
		return
	}

	pokes := make([]*Pokemon_info, cfg.N)
	wg := sync.WaitGroup{}
	for i := 0; i < cfg.N; i++ {
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
	http.HandleFunc("/poke/table", PokeTableHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
