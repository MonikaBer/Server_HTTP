package main
import (
	"io/ioutil"
	"fmt"
	"strconv"
	"encoding/json"
	"net/http"
	"log"
	"time"
	"sync"
	"github.com/go-chi/chi"
)

const TIMEOUT = time.Duration(5 * time.Second)  //timeout for http.Get()

type PageDescription struct {
	Id uint32 `json:"id"`
	Url string `json:"url"`
	Interval uint32 `json:"interval"`
}

type PageEvent struct {
	Response string `json:"response"`
	Duration float64 `json:"duration"`
	Created_at float64 `json:"created_at"`
}

type PageHistory struct {
	Events []PageEvent
	IfPageExists bool
	Interval uint32
	Mtx sync.Mutex 
}

type Storage struct {
	Descriptions map[uint32]PageDescription
	Histories map[uint32]*PageHistory
	NextFreeKey uint32
}

type Tmp struct {
	Url string `json:"url"`
	Interval uint32 `json:"interval"`
}

type Id struct {
	Id uint32 `json:"id"`
}

func newStorage() *Storage {
	var storage Storage
	storage.Descriptions = make(map[uint32]PageDescription)
	storage.Histories = make(map[uint32]*PageHistory)
	storage.NextFreeKey = 0
	return &storage
}

func isPageInStorage(descriptions map[uint32]PageDescription, url string) (uint32, bool) {
	for _, v := range(descriptions) {
		if (v.Url == url) {
			return v.Id, true 
		}
	}
	return 0, false
}


func execWorker(url string, interval uint32, pageHistory *PageHistory) {
	client := http.Client{
		Timeout : TIMEOUT,
	}
	var response []byte

	for true {
		if (*pageHistory).IfPageExists == false {
			return
		}
		
		_created_at := fmt.Sprintf("%.5f", float64(time.Now().UnixNano()) / float64(10e9))
		created_at, err := strconv.ParseFloat(_created_at, 64);
		if err != nil {
			log.Println("Float parsing error during time processing!")
			continue
		}

		time_start := time.Now()
		resp, err := client.Get(url)
		time_end := time.Since(time_start)
		if err != nil {
			log.Println("Web page uploading error - timeout!")
			response = nil
		} else {
			response, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("Body reading error!")
				continue
			}
		}
		defer func() {
			er := resp.Body.Close()
			if er != nil {
				log.Println("Response body closing error!")
			}
		}()

		_duration := fmt.Sprintf("%.3f", float64(time_end) / float64(10e9))
		duration, err := strconv.ParseFloat(_duration, 64);
		if err != nil {
			log.Println("Float parsing error during time processing!")
			continue
		}

		event := PageEvent{string(response), float64(duration), float64(created_at)}
		(*pageHistory).Mtx.Lock()
		(*pageHistory).Events = append((*pageHistory).Events, event)
		interval = (*pageHistory).Interval
		(*pageHistory).Mtx.Unlock()

		time.Sleep(time.Duration(interval) * time.Second)
	}
}


func getPageHistory(storage *Storage, w http.ResponseWriter, r *http.Request) {
	idInString := chi.URLParam(r, "id")
	id, err :=  strconv.Atoi(idInString)
	if err != nil {
		log.Println("String to integer conversion error!")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	history, er := storage.Histories[uint32(id)]
	if er == false {
		log.Println("Web page not found!")
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	} 

	(*history).Mtx.Lock()
	eventsJson, err := json.Marshal((*history).Events)
	if err != nil {
		log.Println("History encoding to JSON error!")
		http.Error(w, err.Error(), http.StatusInternalServerError) 
		return
	}
	(*history).Mtx.Unlock()

	w.Write(eventsJson)
}

func getAllDescriptions(storage *Storage, w http.ResponseWriter, r *http.Request) {
	var descriptions = []PageDescription{}
	for _, v := range(storage.Descriptions) {
		descriptions = append(descriptions,v)
	}

	descriptionsJson, err := json.Marshal(descriptions)
	if err != nil {
		log.Println("Pages' descriptions encoding to JSON error!")
		http.Error(w, err.Error(), http.StatusInternalServerError) 
		return
	}

	w.Write(descriptionsJson)
}

func postWebPage(storage *Storage, w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10e6)  //max size of body in POST 

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Request entity too large! Error!")
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		return
	}
	defer func() {
		err = r.Body.Close()
		if err != nil {
			log.Println("Request body closing error!")
			err = nil
		}
	}()

	var requestBody Tmp
	err = json.Unmarshal(b, &requestBody)
	if err != nil {
		log.Println("Request body in POST method unmarshalling error!")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if requestBody.Interval == 0 {
		log.Println("Invalid request body - lack of Interval value! Error!")
		http.Error(w, "bad request, requestBody.Interval == 0", http.StatusBadRequest)
		return
	}
	
	var pageDescription PageDescription
	pageDescription.Url = requestBody.Url
	pageDescription.Interval = requestBody.Interval

	var tmp Id
	id, er := isPageInStorage(storage.Descriptions, requestBody.Url)
	if er == true {
		pageDescription.Id = id
		storage.Descriptions[id] = pageDescription  //update web page description in storage
		
		(*storage.Histories[id]).Mtx.Lock()
		(*storage.Histories[id]).Interval = requestBody.Interval
		(*storage.Histories[id]).Mtx.Unlock()

		tmp.Id = id
	} else {
		pageDescription.Id = storage.NextFreeKey
		storage.Descriptions[storage.NextFreeKey] = pageDescription  //add new web page description to storage

		var mtx sync.Mutex
		var events []PageEvent
		pageHistory := PageHistory{events, true, pageDescription.Interval, mtx} 
		storage.Histories[storage.NextFreeKey] = &pageHistory  //add new web page history to storage

		go execWorker(pageDescription.Url, pageDescription.Interval, storage.Histories[storage.NextFreeKey])
		storage.NextFreeKey++
	
		tmp.Id = pageDescription.Id
	}

	result, err := json.Marshal(&tmp)
	if err != nil {
		log.Println("Marshalling error!")
	}
	w.Write(result)
}

func deleteWebPage(storage *Storage, w http.ResponseWriter, r *http.Request) {
	idInString := chi.URLParam(r, "id")
	idInInt, err :=  strconv.Atoi(idInString)
	if err != nil {
		log.Println("String to integer conversion error!")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := uint32(idInInt)
	pageDescription, er := storage.Descriptions[id]
	if er == false {
		log.Println("Web page to delete is Not Found")
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}

	var tmp Id
	tmp.Id = pageDescription.Id
	result, err := json.Marshal(&tmp)
	
	delete(storage.Descriptions, id)  //delete web page description from descriptions
	(*storage.Histories[id]).Mtx.Lock()
	(*storage.Histories[id]).IfPageExists = false
	(*storage.Histories[id]).Mtx.Unlock()
	delete(storage.Histories, id)     //delete web page history from history

	if err != nil {
		log.Println("Marshalling error!")
		http.Error(w, err.Error(), http.StatusInternalServerError) 
		return
	}
	w.Write(result)
}


func main() {
	fmt.Println("Starting server on port :8080...")

	storage := newStorage()

	r := chi.NewRouter()
	r.Route("/api", func(r chi.Router) {
		r.Route("/fetcher", func(r chi.Router) {
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/history", func(w http.ResponseWriter, r *http.Request)  { getPageHistory(storage, w, r) })
			})
			r.Get("/", func(w http.ResponseWriter, r *http.Request) { getAllDescriptions(storage, w, r) })
			r.Post("/", func(w http.ResponseWriter, r *http.Request) { postWebPage(storage, w, r) })
			r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) { deleteWebPage(storage, w, r) })
		})
	})

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("Server starting error!")
	}
}